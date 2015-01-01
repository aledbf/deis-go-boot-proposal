package commons

import (
	"bytes"
	"crypto/rand"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/deis/go-boot-proposal/logger"
)

const (
	networkWaitTime time.Duration = 5 * time.Second
)

// Getopt return the value of and environment variable or a default
func Getopt(name, dfault string) string {
	value := os.Getenv(name)
	if value == "" {
		logger.Log.Debugf("returning default value \"%s\" for key \"%s\"", dfault, name)
		value = dfault
	}
	return value
}

// CopyFile copy a file from <src> to <dst>
func CopyFile(dst, src string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func StartServiceCommand(command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Start()

	if err != nil {
		logger.Log.Printf("an error ocurred executing command: %v", err)
		os.Exit(1)
	}

	err = cmd.Wait()
	logger.Log.Printf("command finished with error: %v", err)
}

func RunCommand(command string, args []string) string {
	cmd := exec.Command(command, args...)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()

	if err != nil {
		logger.Log.Printf("an error ocurred executing command: %v", err)
		os.Exit(1)
	}

	return stdout.String()
}

func WaitForLocalConnection(startedChan chan bool, protocol string, testPort string) {
	for {
		_, err := net.DialTimeout(protocol, "127.0.0.1:"+testPort, networkWaitTime)
		if err == nil {
			startedChan <- true
			break
		}
	}
}

// BuildCommandFromString parses a string containing a command and multiple
// arguments and returns a valid tuple to pass to exec.Command
func BuildCommandFromString(input string) (string, []string) {
	command := strings.Split(input, " ")

	if len(command) > 1 {
		return command[0], command[1:]
	}

	return command[0], []string{}
}

func RandomKey() string {
	size := 64
	alphanum := "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	var bytes = make([]byte, size)
	rand.Read(bytes)
	for i, b := range bytes {
		bytes[i] = alphanum[b%byte(len(alphanum))]
	}
	return string(bytes)
}
