package kapetanios

import (
	"bytes"
	"log"
	"os/exec"
)

func main() {
	updateCommand := "yum update -y"
	cmd := exec.Command("/bin/bash", "-c", updateCommand)

	var stdoutBuf, stderrBuf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdoutBuf, &stderrBuf

	err := cmd.Run()
	if err != nil {
		log.Println("failed to fetch repository updates", err)
	}

	log.Println("stdoutBuf:", stdoutBuf.String())
	log.Println("stderrBuf:", stderrBuf.String())
}
