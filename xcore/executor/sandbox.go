package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/talatmursalin/ekshunno-executor/customenums"
	"github.com/talatmursalin/ekshunno-executor/models"
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/exutils"
)

type SandboxExecutor struct {
	compilerSettings compilers.Compiler
	limits           models.Limit
	src              string
	inputFileName    string
	outputFileName   string
	dir              string
	outDir           string
	ctx              context.Context
	client           *client.Client
	container        container.ContainerCreateCreatedBody
}

func goBoomOnError(msg string, err error) {
	if err != nil {
		log.Panicf("%s:%s", msg, err)
	}
}

func (sdb *SandboxExecutor) createConatiner() {
	sdb.ctx = context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	goBoomOnError("Failed to create docker client", err)
	sdb.client = cli
	resp, err := sdb.client.ContainerCreate(
		sdb.ctx,
		&container.Config{
			Image:           sdb.compilerSettings.GetImageName(),
			NetworkDisabled: true,
			Tty:             true,
		},
		&container.HostConfig{
			Mounts: []mount.Mount{
				{
					Type:     mount.TypeBind,
					Source:   sdb.dir,
					Target:   sdb.dir,
					ReadOnly: true,
				},
			},
			Resources: container.Resources{
				CPUCount:   1,
				Memory:     int64(sdb.limits.MemoryLimit * 1024 * 1024), // Memory limit (in bytes)
				MemorySwap: int64(sdb.limits.MemoryLimit * 1024 * 1024),
			},
		},
		nil, nil, "")
	goBoomOnError("Failed to create docker container", err)
	sdb.container = resp
	err = sdb.client.ContainerStart(sdb.ctx, resp.ID, types.ContainerStartOptions{})
	goBoomOnError("Failed to start docker container", err)
}

func (sdb *SandboxExecutor) runInsideDocker(cmds []string) models.ExecResult {
	log.Printf("Exec : %s", cmds)
	conf := types.ExecConfig{
		AttachStdout: false,
		AttachStderr: true,
		Detach:       true,
		Cmd:          cmds,
	}
	execID, _ := sdb.client.ContainerExecCreate(sdb.ctx, sdb.container.ID, conf)

	config := types.ExecStartCheck{}
	resp, err := sdb.client.ContainerExecAttach(sdb.ctx, execID.ID, config)
	goBoomOnError("Failed to attach ExecAttachment", err)
	defer resp.Close()

	err = sdb.client.ContainerExecStart(sdb.ctx, execID.ID, types.ExecStartCheck{})
	goBoomOnError("Failed to start docker exec", err)

	// read the output
	execResult := models.ExecResult{}
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		goBoomOnError("Failed to read docker exit status", err)
		break

	case <-sdb.ctx.Done():
		goBoomOnError("Failed to close docker exec channel", sdb.ctx.Err())
		return execResult
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	goBoomOnError("Failed to read docker stdout", err)
	stderr, err := ioutil.ReadAll(&errBuf)
	goBoomOnError("Failed to read docker stderr", err)

	res, err := sdb.client.ContainerExecInspect(sdb.ctx, execID.ID)
	goBoomOnError("Failed to inspect docker exec", err)

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult
}

func (sdb *SandboxExecutor) absoluteSrcPath() string {
	return filepath.Join(sdb.dir, sdb.compilerSettings.GetSourceName())
}

func (sdb *SandboxExecutor) copySrcToOutDir() {
	cpyCmd := fmt.Sprintf("cp %s %s", sdb.absoluteSrcPath(), sdb.outDir)
	cmds := []string{"bash", "-c", cpyCmd}
	sdb.runInsideDocker(cmds)
}

func (sdb *SandboxExecutor) writeInput(input string) {
	path := filepath.Join(sdb.dir, sdb.inputFileName)
	exutils.WriteFile(path, input)
}

func (sdb *SandboxExecutor) createLocalEnv() {
	exutils.CreateLocalDir(sdb.dir)
	// copy src file to work dir
	err := exutils.WriteFile(sdb.absoluteSrcPath(), sdb.src)
	goBoomOnError("Failed to write file", err)
}

func (sdb *SandboxExecutor) createDockerEnv() {
	// create docker directory
	cmds := []string{"bash", "-c", fmt.Sprintf("mkdir %s", sdb.outDir)}
	sdb.runInsideDocker(cmds)
}

func (sdb *SandboxExecutor) stopAndRemoveContainer() error {

	if err := sdb.client.ContainerStop(sdb.ctx, sdb.container.ID, nil); err != nil {
		log.Printf("Unable to stop container %s: %s", sdb.container.ID, err)
	}

	removeOptions := types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}

	if err := sdb.client.ContainerRemove(sdb.ctx, sdb.container.ID, removeOptions); err != nil {
		log.Printf("Unable to remove container: %s", err)
		return err
	}

	return nil
}

func (sdb *SandboxExecutor) downloadOutput() string {

	outputFilePath := filepath.Join(sdb.outDir, sdb.outputFileName)
	tarStream, _, err := sdb.client.CopyFromContainer(sdb.ctx, sdb.container.ID, outputFilePath)
	goBoomOnError("Failed to copy from container", err)
	tr := tar.NewReader(tarStream)
	if _, err := tr.Next(); err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(tr)

	return buf.String()
}

func (sdb *SandboxExecutor) Compile() models.Result {
	srcDir := sdb.dir
	outDir := sdb.outDir
	if sdb.compilerSettings.IsInterpreter() {
		sdb.copySrcToOutDir()
		srcDir = outDir
	}
	compileCmd := sdb.compilerSettings.GetCompileCommand(srcDir, outDir)
	fmtCmd := fmt.Sprintf("cd %s && timeout -s KILL %f %s", srcDir, sdb.limits.TimeLimit, compileCmd)
	cmds := []string{"bash", "-c", fmtCmd}
	compilerResult := sdb.runInsideDocker(cmds)
	log.Printf("compile cmd exit code: %d", compilerResult.ExitCode)
	res := models.Result{
		Verdict: customenums.OK,
		Time:    0,
		Memory:  0,
		Output:  compilerResult.StdOut,
	}
	if compilerResult.ExitCode != 0 {
		log.Printf("compile error: %s - %s", compilerResult.StdOut, compilerResult.StdErr)
		switch compilerResult.ExitCode {
		case 4:
			res.Verdict = customenums.IE
			res.Output = "Insufficient memory"
		default:
			res.Verdict = customenums.CE
			res.Output = compilerResult.StdErr
		}
	}
	return res
}

func determineVerdict(result models.ExecResult) customenums.VerdictEnum {
	switch result.ExitCode {
	case 0:
		return customenums.OK
	case 124:
		return customenums.TLE
	case 137:
		return customenums.MLE
	case 153:
		return customenums.OLE
	default:
		return customenums.RE
	}
}

func getExecutionTime(result string) float32 {
	pattern := "real\t([0-9]+)m([0-9]+\\.[0-9]{3})s\n"
	regex := regexp.MustCompile(pattern)
	match := regex.FindStringSubmatch(result)
	if len(match) == 0 {
		return 0
	}
	t, err := strconv.ParseFloat(match[len(match)-1], 32)
	if err != nil {
		return 0
	}
	return float32(t)
}

func (sdb *SandboxExecutor) getContainerStats() string {
	stats, err := sdb.client.ContainerStatsOneShot(sdb.ctx, sdb.container.ID)
	if err != nil {
		log.Printf("Failed to read conatiner stat: %s", err)
	}
	defer stats.Body.Close()
	content, _ := ioutil.ReadAll(stats.Body)
	return string(content)
}

func getMemoryFromStat(result string) float32 {
	pattern := "\"max_usage\":([0-9]+)"
	regex := regexp.MustCompile(pattern)
	match := regex.FindString(result)
	if len(match) == 0 {
		return 0
	}
	mStr := strings.Split(match, ":")[1]
	m, err := strconv.ParseFloat(mStr, 32)
	if err != nil {
		return 0
	}
	return float32(m) / (1024 * 1024)
}

func processedError(stderr, replace string) string {
	// log.Printf("-->%s\n -->%s\n", stderr, replace)
	err := strings.Split(stderr, "real\t")[0]
	for _, val := range strings.Split(replace, " ") {
		err = strings.ReplaceAll(err, val, "")
	}
	return err
}

func (sdb *SandboxExecutor) prepareExecuteCommand() string {
	command := fmt.Sprintf("set -o pipefail && ulimit -f %d && cd %s && time timeout %f %s < %s > %s",
		int64(sdb.limits.OutputLimit*1024), //kb
		sdb.outDir,
		sdb.limits.TimeLimit,
		sdb.compilerSettings.GetExecuteCommand(sdb.outDir),
		filepath.Join(sdb.dir, sdb.inputFileName),
		sdb.outputFileName,
	)
	return command
}

func (sdb *SandboxExecutor) Execute(io string) models.Result {
	defer sdb.Clear()
	sdb.writeInput(io)
	exeCmd := sdb.prepareExecuteCommand()
	cmds := []string{"bash", "-c", exeCmd}
	res := sdb.runInsideDocker(cmds)
	log.Printf("execute cmd exit code: %d", res.ExitCode)
	// log.Printf("execute stderr: %s", res.StdErr)
	output := ""
	if res.ExitCode != 0 {
		output = processedError(res.StdErr, exeCmd) + "\n\n"
	}
	output += sdb.downloadOutput()
	cStat := sdb.getContainerStats()
	result := models.Result{
		Verdict: determineVerdict(res),
		Time:    getExecutionTime(res.StdErr),
		Memory:  getMemoryFromStat(cStat),
		Output:  output,
	}
	return result
}

func (sdb *SandboxExecutor) Clear() {
	sdb.stopAndRemoveContainer()
	exutils.DeleteLocalDir(sdb.dir)
}

func NewSandboxExecutor(src string, sett compilers.Compiler, limits models.Limit) *SandboxExecutor {
	sdb := SandboxExecutor{
		compilerSettings: sett,
		limits:           limits,
		src:              src,
		inputFileName:    "input.in",
		outputFileName:   "output.out",
		dir:              exutils.TempDirName("es_"),
		outDir:           exutils.TempDirName("es_out_"),
	}
	sdb.createLocalEnv()
	sdb.createConatiner()
	sdb.createDockerEnv()
	return &sdb
}
