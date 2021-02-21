package main

import (
	"flag"                                    // implements command-line flag parsing	
	"github.com/StarForger/neb-zap-config"    // logging config
	"go.uber.org/zap"                         // structured, leveled logging
	"io"                                      // basic interfaces to I/O primitives	
	"os"                                      // platform-independent interface to operating system functionality
	"os/exec"                                 // runs external commands
	"os/signal"                               // access to incoming signals	
	"syscall"                                 // interface to the low-level operating system primitives
	"time"                                    // functionality for measuring and displaying time
)

const name = "mc-wrapper"

func main() { 
	// Arguments	
  debug         := *flag.Bool("v", false, "Enable debug logging")
  detachStdin   := *flag.Bool("d", false, "Don't forward stdin and allow process to be put in background")
  shell         := *flag.String("shell", "", "")
  stopDuration  := *flag.Duration("s", 0, "Duration to wait after sending the 'stop' command")

  // Parse Flags
  flag.Parse()  

	// Logs
	logger := zap-config.LoggerInfo()
	if debug {
		logger = zap-config.LoggerDebug()
	}
	defer logger.Sync()
	logger = logger.Named(name)	

	// Command Error Check
	if flag.NArg() < 1 {
		logger.Fatal("Missing executable arguments")
	}

	// external command being prepared  
  var cmd *exec.Cmd
	if shell != "" {
    cmd = exec.Command(shell, flag.Args()...)
  } else {
    cmd = exec.Command(flag.Arg(0), flag.Args()[1:]...)
  }  

  // Stdin
  stdin, err := cmd.StdinPipe()
  if err != nil {
    logger.Error("Unable to pipe stdin", zap.Error(err))
  }

  // Stdout
  stdout, err := cmd.StdoutPipe()
  if err != nil {
    logger.Error("Unable to pipe stdout", zap.Error(err))
  }

  // Stderr
  stderr, err := cmd.StderrPipe()
  if err != nil {
    logger.Error("Unable to pipe stderr", zap.Error(err))
  }

  // Start command
  err = cmd.Start()
  if err != nil {
    logger.Error("Failed to start", zap.Error(err))
  }

  // Relay stdin/out/err between os and command
  if !detachStdin {
    go func() {
      io.Copy(stdin, os.Stdin)
    }()
  }
  go func() {
    io.Copy(os.Stdout, stdout)
  }()
  go func() {
    io.Copy(os.Stderr, stderr)
  }()  

  // initialize a buffered channel of os.signal type
  osSignalChannel := make(chan os.Signal, 1)
  // initialize a buffered channel for exit
  cmdExitChannel := make(chan int, 1)
  // intercept SIGTERM, send via signalChan
  signal.Notify(signalChan, syscall.SIGTERM)	

  // Wait Goroutine (async)
  go func() {
    // waits for the command to exit    
    if err := cmd.Wait(); err != nil {
      // Type Assert exit code to type *exec.ExitError
      // The value of ok is true if the assertion holds
      if exitErr, ok := err.(*exec.ExitError); ok {
        exitCode := exitErr.ExitCode()
        logger.Warn("executable sub-process failed", zap.Int("exitCode", exitCode))
        cmdExitChannel <- exitCode
      } else {
        logger.Error("executable failed", zap.Error(err))
        cmdExitChannel <- 1
      }
      return
    } else {
      cmdExitChannel <- 0
    }
  }()

  // Wait for channel output
  for {
    select {
    case <-osSignalChannel:			
      stopViaConsole(logger, stdin)			

      logger.Info("Waiting for completion...")
      if stopDuration > 0 {
        time.AfterFunc(stopDuration, func() {
          logger.Error("Took too long, killing server process")
          err := cmd.Process.Kill()
          if err != nil {
            logger.Error("failed to forcefully kill process")
          }
        })
      }

    case exitCode := <-cmdExitChannel:
      logger.Info("Done")
      os.Exit(exitCode)
    }
  } 
}

func stopViaConsole(logger *zap.Logger, stdin io.Writer) {
	logger.Info("Sending 'stop' to Minecraft server...")
	_, err := stdin.Write([]byte("stop\n"))
	if err != nil {
		logger.Error("failed to write stop command to server console", zap.Error(err))
	}
}
