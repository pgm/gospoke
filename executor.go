package main

import ( 
	"os"
	"fmt"
	"log"
	)

func ExecuteCommand(command string, input string) {

	stdoutRead, stdoutWrite, _ := os.Pipe()
	stdinRead, stdinWrite, _ := os.Pipe()

	attr := &os.ProcAttr{".", nil, []*os.File{stdinRead, stdoutWrite, stdoutWrite}}
	proc, err := os.StartProcess(command, []string{command}, attr)

	stdoutWrite.Close()
	stdinRead.Close()

	if err == nil {

		// create two go-routines.  One for reading, one for writing
		inputBytes := []byte(input)
		go func () {
			_, _ = stdinWrite.Write(inputBytes)
			stdinWrite.Close()
			log.Println("Completed writing to child proccs")
		}()

		go func () { 
			for {
				// allocate a new buffer
				buffer := make([]byte, 1000)

				// read into that buffer
				count, err := stdoutRead.Read(buffer)

				if count > 0 {
					fmt.Printf("output from command: %s\n", buffer[:count])
				}
				
				// if we reached the end, bail from this loop
				if err == os.EOF {
					break
				}
			}
			
			stdoutRead.Close()

			log.Println("Waiting to reap child process")
			// reap child process
			_, _ = proc.Wait(0)
			
//			if error != nil { 
//				fmt.Printf("error=%v\n", error)
//			}
			log.Println("Go routine terminating")
		}()
		
	} else {
		log.Println("Error: "+err.String())
		stdoutRead.Close()
		stdinWrite.Close()
		log.Println("Cleaned up handles after error")
	}
	
//	return err
}
