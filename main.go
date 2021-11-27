package main

import (
	"bytes"
	"flag"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
)

const (
	PIPE_BUFFER_SIZE   = 64                     // explained in note 2
	PIPE_READ_INTERVAL = time.Second            // explained in note 3
	KILL_SIGNAL        = "LOOP_EXIT"            // explained in note 4
	MAIN_LOOP_INTERVAL = time.Millisecond * 100 // explained in note 5
	MAIN_BUFFER_SIZE   = 512                    // explained on line 66
)

var defaultCommand = "echo hello world"

func main() {
	log.Print("program is running now")

	//getting user home directory to use it as default
	home, err := os.UserHomeDir()

	if err != nil {
		log.Fatalf("could not get user home directory : %s", err)
	}

	//setting flags to get user information flags
	pipeName := flag.String("pname", "/tmp/pp", "absolute path(inclding name) of the named pipe you want to use")
	directoryPath := flag.String("working-drectory", home, "working directory of program")
	command := flag.String("cmd", defaultCommand, "actual command to execute")

	flag.Parse() // parsing user args to get flag inputs

	go func() {

		// see note 1 at bottom for more information about this loop
		for {
			file, err := os.OpenFile(*pipeName, os.O_RDONLY, 0600)

			if err != nil {
				log.Fatal(err)
			}

			// see note 2 for more details
			buffer := make([]byte, PIPE_BUFFER_SIZE)

			// see note 3 for why i have added time.sleep() here
			time.Sleep(PIPE_READ_INTERVAL)
			_, err = file.Read(buffer)
			if err != nil {
				log.Printf("could not read pipe %s", err.Error())
			}

			// check note 4 for why strings.Contains instead of ==
			if strings.Contains(string(buffer), KILL_SIGNAL) {
				log.Print("exiting ...")
				os.Exit(0)
			} else {
				log.Printf("incorrect exit input:%s", string(buffer))
			}
			file.Close()
		}
	}()

	for {

		log.Printf("executing command %s", *command)
		buffer := bytes.NewBuffer(make([]byte, MAIN_BUFFER_SIZE))

		cmd := exec.Command("bash", "-c", *command)
		cmd.Stdout = buffer      // for storing output, upto first 512 chars from it, you can increase it increasing value of BUFFER_SIZE constant
		cmd.Dir = *directoryPath // changing directory as some commands might rely on files kept in specific directoy
		err := cmd.Run()
		if err != nil {
			log.Fatal(err)
		}

		log.Print(buffer.String())

		// see note 5 for why adding time.Sleep() here
		time.Sleep(MAIN_LOOP_INTERVAL)
	}
}

/*

Note 1 ---------------------------------------------------------------------------------

> i am using blocking action of reading from empty pipe as blocking mechanism
  to trigger program actions

> repeating the entire logic inside loop because once data is read from
pipe, it needs to be opened again to read new values inside it, otherwise it keeps
reading old vales and does not even blocks execution

Note 2 ---------------------------------------------------------------------------------

this the buffer where data read from buffer will be stored
please keep in mind that your pipe buffer size should be bigger
than length of your KILL_SIGNAL otherwise it will never fit
and hence will never be right on comparison step,
edit value of PIPE_BUFFER_SIZE to match *your* KILL_SIGNAL SIZE

Note 3 ---------------------------------------------------------------------------------

adding sleep option between tries so that it does not completely
gobbles system resources, set it accordingly and try not to make it 0
as its bit risky because in inifinite loop it can crash system,
setting it on 0 is effectively ddosing yourself in case of problem

Note 4 --------------------------------------------------------------------------------

using string.Contains instead of == because output from pipe
as it contains some other characters which might not be consistent OSes
and checking for specific char sequence, which is defined in KILL_SIGNAL
constant defined at top

Note 5 --------------------------------------------------------------------------------

adding sleep option between runs so that it does not completely
gobbles system resources, set it accordingly and try not to make it 0
as its bit risky because in inifinite loop it can crash system,
setting it on 0 is effectively ddosing yourself in case of problem

*/
