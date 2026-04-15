package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"feishu-pipeline/apps/api-go/internal/bootstrap"

	"github.com/spf13/cobra"
)

func newServeCommand() *cobra.Command {
	var configPath string

	command := &cobra.Command{
		Use:   "serve",
		Short: "启动 Gin API 服务",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
			defer stop()

			application, err := bootstrap.NewApplication(ctx, configPath, version)
			if err != nil {
				return err
			}
			defer application.Close()

			go func() {
				<-ctx.Done()
				shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = application.HTTPServer.Shutdown(shutdownCtx)
			}()

			fmt.Fprintf(cmd.OutOrStdout(), "requirement delivery api listening on %s\n", application.HTTPServer.Addr)
			err = application.HTTPServer.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				return err
			}
			return nil
		},
	}

	command.Flags().StringVar(&configPath, "config", "", "配置文件路径，默认读取 config/config.yaml")
	return command
}

func Execute() {
	root := NewRootCommand()
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
