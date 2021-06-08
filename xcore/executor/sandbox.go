package executor

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/talatmursalin/ekshunno-executor/xcore/compilers"
	"github.com/talatmursalin/ekshunno-executor/xcore/utils"
)

type SandboxExecutor struct {
	compilerSettings compilers.Compiler
	limits           utils.Limit
	src              string
	inputFileName    string
	outputFileName   string
	dir              string
	outDir           string
	ctx              context.Context
	client           *client.Client
	container        container.ContainerCreateCreatedBody
}

func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

func (sdb *SandboxExecutor) createConatiner() {
	sdb.ctx = context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	checkError(err)
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
	checkError(err)
	sdb.container = resp
	err = sdb.client.ContainerStart(sdb.ctx, resp.ID, types.ContainerStartOptions{})
	checkError(err)
}

func (sdb *SandboxExecutor) runInsideDocker(cmds []string) utils.ExecResult {
	fmt.Println("Exec : ", cmds)
	conf := types.ExecConfig{
		AttachStdout: false,
		AttachStderr: true,
		Detach:       true,
		Cmd:          cmds,
	}
	execID, _ := sdb.client.ContainerExecCreate(sdb.ctx, sdb.container.ID, conf)

	config := types.ExecStartCheck{}
	resp, err := sdb.client.ContainerExecAttach(sdb.ctx, execID.ID, config)
	checkError(err)
	defer resp.Close()

	err = sdb.client.ContainerExecStart(sdb.ctx, execID.ID, types.ExecStartCheck{})
	checkError(err)

	// read the output
	execResult := utils.ExecResult{}
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err := <-outputDone:
		checkError(err)
		break

	case <-sdb.ctx.Done():
		checkError(sdb.ctx.Err())
		return execResult
	}

	stdout, err := ioutil.ReadAll(&outBuf)
	checkError(err)
	stderr, err := ioutil.ReadAll(&errBuf)
	checkError(err)

	res, err := sdb.client.ContainerExecInspect(sdb.ctx, execID.ID)
	checkError(err)

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
	utils.WriteFile(path, input)
}

func (sdb *SandboxExecutor) createLocalEnv() {
	utils.CreateLocalDir(sdb.dir)
	// copy src file to work dir
	err := utils.WriteFile(sdb.absoluteSrcPath(), sdb.src)
	checkError(err)
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
	checkError(err)
	tr := tar.NewReader(tarStream)
	if _, err := tr.Next(); err != nil {
		panic(err)
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(tr)

	return buf.String()
}

func (sdb *SandboxExecutor) Compile() utils.Result {
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

	res := utils.Result{
		Verdict: utils.OK,
		Time:    0,
		Memory:  0,
		Output:  compilerResult.StdOut,
	}
	if compilerResult.ExitCode != 0 {
		res.Verdict = utils.CE
		res.Output = compilerResult.StdErr
	}
	return res
}

func determineVerdict(result utils.ExecResult) utils.VerdictEnum {
	switch result.ExitCode {
	case 0:
		return utils.OK
	case 124:
		return utils.TLE
	case 137:
		return utils.MLE
	case 153:
		return utils.OLE
	default:
		return utils.RE
	}
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

func (sdb *SandboxExecutor) Execute(io string) utils.Result {
	sdb.writeInput(io)
	exeCmd := sdb.prepareExecuteCommand()
	cmds := []string{"bash", "-c", exeCmd}
	res := sdb.runInsideDocker(cmds)

	fmt.Println("Exit: ", res.StdErr)

	sdb.downloadOutput()
	result := utils.Result{
		Verdict: determineVerdict(res),
		Time:    0.5,
		Memory:  232.07,
		Output:  sdb.downloadOutput(),
	}
	return result
}

func (sdb *SandboxExecutor) Clear() {
	sdb.stopAndRemoveContainer()
	utils.DeleteLocalDir(sdb.dir)
}

func NewSandboxExecutor(src string, sett compilers.Compiler, limits utils.Limit) *SandboxExecutor {
	sdb := SandboxExecutor{
		compilerSettings: sett,
		limits:           limits,
		src:              src,
		inputFileName:    "input.in",
		outputFileName:   "output.out",
		dir:              utils.TempDirName("soj_"),
		outDir:           utils.TempDirName("soj_out_"),
	}
	sdb.createLocalEnv()
	sdb.createConatiner()
	sdb.createDockerEnv()
	return &sdb
}
