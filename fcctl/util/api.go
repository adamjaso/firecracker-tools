package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/firecracker-microvm/firecracker-go-sdk"
	"github.com/firecracker-microvm/firecracker-go-sdk/client/models"
)

const (
	CommandStop  = "stop"
	CommandStart = "start"
	CommandCurl  = "curl"
)

var (
	ErrUnknownCommand = errors.New("unknown command")
	ErrEmptyCurlArgs  = errors.New("empty curl args")
	ErrSocket         = errors.New("socket error")
)

type FcReq struct {
	Method string
	Path   string
	Body   io.Reader
	Res    any
}

func CheckUnixSocket(sockF string) error {
	s, err := os.Stat(sockF)
	if err != nil {
		return fmt.Errorf("%w: %s not found", ErrSocket, sockF)
	} else if s.Mode()|os.ModeSocket == 0 {
		return fmt.Errorf("%w: %s not a socket", ErrSocket, sockF)
	}
	conn, err := net.Dial("unix", sockF)
	if err != nil {
		if os.IsPermission(err) {
			return err
		}
		// socket is not open
		return fmt.Errorf("%w: open error %v", ErrSocket, err)
	}
	_ = conn.Close()
	// socket open
	return nil
}

func CallFirecracker(ctx context.Context, sockF string, req FcReq) error {
	conn, err := net.Dial("unix", sockF)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSocket, err)
	}
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return conn, nil
			},
		},
	}
	log.Printf("%s %s", req.Method, req.Path)
	hreq, err := http.NewRequestWithContext(ctx, req.Method, "http://_"+req.Path, req.Body)
	if err != nil {
		return err
	}
	hres, err := client.Do(hreq)
	if err != nil {
		return err
	}
	log.Printf("HTTP %d", hres.StatusCode)
	defer hres.Body.Close()
	switch t := req.Res.(type) {
	case *os.File:
		if _, err := io.Copy(t, hres.Body); err != nil {
			return err
		}
	default:
		if req.Res != nil {
			return json.NewDecoder(hres.Body).Decode(req.Res)
		}
	}
	return nil
}

func RunFirecrackerCommand(ctx context.Context, sockF, command string, args ...string) error {
	var req FcReq
	switch command {
	case "stop":
		body, _ := json.Marshal(&models.InstanceActionInfo{
			ActionType: firecracker.String(models.InstanceActionInfoActionTypeSendCtrlAltDel),
		})
		req = FcReq{
			Method: http.MethodPut,
			Path:   "/actions",
			Body:   bytes.NewBuffer(body),
			Res:    os.Stdout,
		}
	case "status":
		req = FcReq{
			Method: http.MethodGet,
			Path:   "/",
			Res:    os.Stdout,
		}
	case "curl":
		if len(args) < 2 {
			return fmt.Errorf("%w: curl METHOD PATH", ErrEmptyCurlArgs)
		}
		method := args[0]
		path := args[1]
		req = FcReq{
			Method: method,
			Path:   path,
			Res:    os.Stdout,
		}
		if method == http.MethodPut || method == http.MethodPatch || method == http.MethodPost {
			body := &bytes.Buffer{}
			_, _ = io.Copy(body, os.Stdin)
			req.Body = body
		}
	default:
		return ErrUnknownCommand
	}
	return CallFirecracker(ctx, sockF, req)
}

func GetInstanceInfo(ctx context.Context, sockF string) (models.InstanceInfo, error) {
	info := models.InstanceInfo{}
	req := FcReq{
		Method: http.MethodGet,
		Path:   "/",
		Res:    &info,
	}
	return info, CallFirecracker(ctx, sockF, req)
}

func WaitUntilState(ctx context.Context, sockF, state string) error {
loop:
	for {
		if info, err := GetInstanceInfo(ctx, sockF); err != nil {
			if state == models.InstanceInfoStateNotStarted && errors.Is(err, ErrSocket) {
				return nil
			}
			return err
		} else if *info.State == state {
			break loop
		}
		time.Sleep(time.Second)
	}
	return nil
}
