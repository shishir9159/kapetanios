package tui

import (
	"errors"
	"fmt"
	"github.com/chzyer/readline"
	"github.com/spf13/cobra"
	"io"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
)

var options struct {
	printVersion bool
	insecure     bool
	origin       string
}

func main() {

	rootCmd := &cobra.Command{
		Use:   "ws URL",
		Short: "websocket tool",
		Run:   root,
	}
	rootCmd.Flags().StringVarP(&options.origin, "origin", "o", "", "websocket origin")
	rootCmd.Flags().BoolVarP(&options.printVersion, "version", "v", false, "print version")
	rootCmd.Flags().BoolVarP(&options.insecure, "insecure", "k", false, "skip ssl certificate check")

	err := rootCmd.Execute()
	if err != nil {
		return
	}
}

func root(cmd *cobra.Command, args []string) {

	if len(args) != 1 {
		err := cmd.Help()
		if err != nil {
			return
		}
		os.Exit(1)
	}

	dest, err := url.Parse(args[0])
	if err != nil {
		_, err := fmt.Fprintln(os.Stderr, err)
		if err != nil {
			return
		}
		os.Exit(1)
	}

	var origin string
	if options.origin != "" {
		origin = options.origin
	} else {
		originURL := *dest
		if dest.Scheme == "wss" {
			originURL.Scheme = "https"
		} else {
			originURL.Scheme = "http"
		}
		origin = originURL.String()
	}

	var historyFile string
	cmdUser, err := user.Current()
	if err == nil {
		historyFile = filepath.Join(cmdUser.HomeDir, ".ws_history")
	}

	err = connect(dest.String(), origin, &readline.Config{
		Prompt:      "> ",
		HistoryFile: historyFile,
	}, options.insecure)

	if err != nil {
		_, er := fmt.Fprintln(os.Stderr, err)
		if er != nil {
			return
		}
		if er != io.EOF && !errors.Is(er, readline.ErrInterrupt) {
			os.Exit(1)
		}
	}
}
