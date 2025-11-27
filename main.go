package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cppCode := "#include <iostream>\nint main(){ std::cout <<\"Hello Forge\";return 0;}"
	err := os.WriteFile("submission.cpp", []byte(cppCode), 0644) //Go demands this.
	//What is it? This is a Unix Permission Code (Octal).
	//6 (Owner): Read + Write.
	//4 (Group): Read only.
	//4 (Everyone): Read only.
	if err != nil {
		fmt.Println("Error writing file: ", err)
		return
	}
	fmt.Println("File Saved Sucessfully!")
	cmd := exec.Command("g++", "submission.cpp", "-o", "submission.exe")
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Compilation Failed!")
		fmt.Println(string(output))
		return
	}
	fmt.Println("Binary Built Sucessfully!")
	// ... Binary Built Successfully ...
	fmt.Println("Running the Binary...")
	//Now running the .Exe file
	runCmd := exec.Command("./submission.exe")
	//Run it now
	runOutput, runErr := runCmd.CombinedOutput()
	//Runtime Errors
	if runErr != nil {
		fmt.Println("Runtime error!")
		fmt.Println(string(runOutput))
	}
	//Success
	fmt.Println("Output Sucessful: ")
	fmt.Println(string(runOutput))
}
