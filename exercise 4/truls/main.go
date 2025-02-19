package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	var cmd *exec.Cmd
	exePath, _ := os.Getwd()

	fmt.Println("path: %s", exePath)
	cmd = exec.Command("kgx", "--", "bash", "-c", fmt.Sprintf("./test"))
//	cmd = exec.Command("kgx", "cd ", exePath)
//	cmd = exec.Command("kgx", fmt.Sprintf("cd %s", exePath))
//	cmd = exec.Command("kgx", "--", "bash", "-c", "cd ", exePath, "&& ./test -role=backup")
//	cmd = exec.Command("kgx", "--", "bash", "-c" "cd ", exePath, ";", "./test -role=backup")
 //cmd = exec.Command("man less")
	fmt.Println(cmd)
	err := cmd.Start()
	if err != nil {
		fmt.Println("Error, ", err)
	}
}
