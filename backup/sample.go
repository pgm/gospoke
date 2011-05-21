package main

import (
	"fmt"
	"bytes"
	"os"
	)


const ( PROC_DATA_ARRIVED int = iota 
        PROC_DIED ) 

type Update struct {
	owner *Process
	op int
	buffer []byte
}

type Process struct { 
	name string
	binary string
	args []string
}

func FilterUpdates (c chan *Update, dest chan *Update) {
	procBuffer := make(map[*Process] *bytes.Buffer)
	
	for {
		update := <-c
		proc := update.owner

		lineBuffer, found := procBuffer[proc]
		if !found {
			lineBuffer = bytes.NewBufferString("")
			procBuffer[proc] = lineBuffer
		}

		if update.op == PROC_DATA_ARRIVED {

			b := update.buffer
			for {
				index := bytes.IndexByte(b, '\n')
				if index >= 0 {
					lineBuffer.Write(b[:index])
					fmt.Printf("%s: %s\n", proc.name, lineBuffer.Bytes())
					lineBuffer.Truncate(0)
					b = (b[index+1:])
				} else {
					lineBuffer.Write(b)
					break
				}
			}
		} else if update.op == PROC_DIED {
			fmt.Printf("%s: %s\n", proc.name, lineBuffer.Bytes())
			procBuffer[proc] = nil, false
		}
		
		dest <- update
	}
}


func (p *Process) Start(c chan *Update) os.Error {
	stdoutRead, stdoutWrite, _ := os.Pipe()

	attr := &os.ProcAttr{".", nil, []*os.File{nil, stdoutWrite, stdoutWrite}}
	proc, err := os.StartProcess(p.binary, p.args, attr)

	if err == nil {
		stdoutWrite.Close()
		go func () { 
			for {
				// allocate a new buffer
				buffer := make([]byte, 10)

				// read into that buffer
				count, err := stdoutRead.Read(buffer)

				if count > 0 {
					c <- &Update{p, PROC_DATA_ARRIVED, buffer[:count]} 
				}
				
				// if we reached the end, bail from this loop
				if err == os.EOF {
					break
				}
			}
			
			// the handle has been closed, so just wait for the process to terminate
			stdoutRead.Close()
			proc.Wait(0)
			
			// send back notification of the process termination
			c <- &Update{p, PROC_DIED, nil} 
		}()
	} else {
		stdoutRead.Close()
		stdoutWrite.Close()
	}
	
	return err
}

func method() {
	fmt.Printf("methodcalled\n")
}


func main() {
	procUpdateChannel1 := make(chan *Update, 100)
	procUpdateChannel2 := make(chan *Update, 100)

	go FilterUpdates(procUpdateChannel1, procUpdateChannel2)

	p := &Process{"ls", "/bin/ls", []string {"/bin/ls", "-l"}}

	err := p.Start(procUpdateChannel1)
	if err != nil {	
		fmt.Printf("err=%v\n", err);
	}

	for {
		update := <-procUpdateChannel2
		if(update.op == PROC_DIED)  {
			break;
		}
	}
	
	fmt.Printf("proc died\n");
}
