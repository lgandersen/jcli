package cli

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	Openapi "jcli/client"

	"github.com/gorilla/websocket"
	"github.com/spf13/cobra"
)

const url = "http://localhost:8085/"
const ws = "ws://localhost:8085/containers/%s/attach"
const succesful_ws_exit = "websocket: close 1000 (normal): exit:"

func NewContainerCommand() *cobra.Command {
	containerCmd := &cobra.Command{
		Use:                   "container",
		Short:                 "Manage containers",
		Long:                  `Manage containers`,
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("The container main command have been executed")
		},
	}

	containerCmd.AddCommand(NewContainerCreateCommand())
	containerCmd.AddCommand(NewContainerRemoveCommand())
	containerCmd.AddCommand(NewContainerStartCommand())
	containerCmd.AddCommand(NewContainerStopCommand())
	containerCmd.AddCommand(NewContainerListCommand())
	return containerCmd
}

func NewContainerCreateCommand() *cobra.Command {
	config := Openapi.ContainerCreateJSONRequestBody{
		Networks:  &([]string{}),
		Volumes:   &([]string{}),
		Env:       &([]string{}),
		JailParam: &([]string{}),
	}

	var name string

	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new container",
		Long:  `Create a new container`,
		Run: func(cmd *cobra.Command, args []string) {
			RunContainerCreate(cmd, &name, config, args)
		},
	}

	flags := createCmd.Flags()
	flags.StringVar(&name, "name", "", "Assign a name to the container")
	flags.StringSliceVar(config.Networks, "network", []string{}, "Connect a container to a network")
	flags.StringSliceVarP(config.Volumes, "volume", "v", []string{}, "Bind mount a volume to the container")
	flags.StringSliceVarP(config.Env, "env", "e", []string{}, "Set environment variables (e.g. --env FIRST=env --env SECOND=env)")
	flags.StringSliceVarP(config.JailParam, "jailparam", "J", []string{"mount.devfs"}, "Specify a jail parameter (see jail(8) for details)")
	return createCmd
}

func RunContainerCreate(cmd *cobra.Command, name *string, body Openapi.ContainerCreateJSONRequestBody, args []string) {
	response, err := PostContainerCreate(name, body, args)
	if err == nil {
		fmt.Println(response.JSON201.Id)
	}
}

func PostContainerCreate(name *string, body Openapi.ContainerCreateJSONRequestBody, args []string) (*Openapi.ContainerCreateResponse, error) {
	container_cmd := args[1:]
	image := args[0]
	body.Cmd = &container_cmd
	body.Image = &image

	params := Openapi.ContainerCreateParams{}
	if *name != "" {
		params = Openapi.ContainerCreateParams{Name: name}
	}

	client := NewHTTPClient()

	response, err := client.ContainerCreateWithResponse(context.TODO(), &params, body)
	if err != nil {
		fmt.Println("Could not connect to jocker engine daemon: ", err)
		return response, err
	}

	if response.StatusCode() != 201 {
		fmt.Println("Jocker engine returned unsuccesful statuscode: ", response.Status())
		return response, errors.New("non-200 statuscode")
	}
	return response, nil
}

func NewContainerRemoveCommand() *cobra.Command {
	removeCmd := &cobra.Command{
		Use:   "rm",
		Short: "Remove one or more containers",
		Long:  `Remove one or more containers loooong`,
		Run: func(cmd *cobra.Command, args []string) {
			RunContainerRemove(cmd, args)
		},
	}
	return removeCmd
}

func RunContainerRemove(cmd *cobra.Command, args []string) {
	response, err := PostContainerRemove(args)
	if err == nil {
		fmt.Println(response.JSON200.Id)
	}
}

func PostContainerRemove(args []string) (*Openapi.ContainerDeleteResponse, error) {
	container_id := args[0]
	client := NewHTTPClient()
	response, _ := client.ContainerDeleteWithResponse(context.TODO(), container_id)
	status_code := response.StatusCode()

	switch {
	case status_code == 200:
		//fmt.Println("succesfully removed container")
		return response, nil
	case status_code == 404:
		return response, errors.New("no such container")
	case status_code == 500:
		return response, errors.New("internal server error")
	default:
		return response, errors.New("unknown status-code received from jocker engine: " + response.Status())
	}
}

func NewContainerStartCommand() *cobra.Command {
	var attach bool
	cmd := &cobra.Command{
		Use:                   "start [OPTIONS] CONTAINER [CONTAINER...]",
		Short:                 "Start one or more stopped containers",
		Long:                  "Start one or more stopped containers. Attach to STDOUT/STDERR if only one container is started",
		Args:                  cobra.MinimumNArgs(1),
		DisableFlagsInUseLine: true,
		Run: func(cmd *cobra.Command, args []string) {
			RunContainerStart(cmd, &attach, args)
		},
	}

	cmd.Flags().BoolVarP(&attach, "attach", "a", true, "Attach STDOUT/STDERR")
	return cmd
}

func RunContainerStart(cmd *cobra.Command, attach *bool, args []string) {
	if *attach {
		if len(args) == 1 {
			StartAndAttachToContainer(attach, args[0])
		} else {
			fmt.Println("When attaching to STDOUT/STDERR only 1 container can be started")
		}
	} else {
		client := NewHTTPClient()
		var container_id string
		for _, container := range args {
			container_id = StartSingleContainer(client, container)
			if container_id != "" {
				fmt.Println(container_id)
			}
		}
	}
}

func StartAndAttachToContainer(attach *bool, container_id string) {
	endpoint := fmt.Sprintf(ws, container_id)
	c, _, err := websocket.DefaultDialer.Dial(endpoint, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	done := make(chan struct{})

	go ListenForWSMessages(done, c)
	StartSingleContainer(NewHTTPClient(), container_id)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			fmt.Println("Interrupted by user")
			TryGracefulWSDisconnectconnect(done, c)
			return
		}
	}
}

func ListenForWSMessages(done chan struct{}, c *websocket.Conn) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if strings.HasPrefix(err.Error(), succesful_ws_exit) {
				fmt.Println(err.Error()[len(succesful_ws_exit):])
			} else {
				fmt.Println("websocket closed unexpectedly:", err.Error())
			}
			return
		}
		msg := string(message)
		if msg[:3] == "ok:" {
			// First message receieved when the ws is succesfully established.
			continue
		}
		if msg[:3] == "io:" {
			fmt.Print(msg[3:])
		}
	}
}

func TryGracefulWSDisconnectconnect(done chan struct{}, c *websocket.Conn) {
	_ = c.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	select {
	case <-done:
	case <-time.After(time.Second):
	}
}

func StartSingleContainer(client *Openapi.ClientWithResponses, container string) string {
	response, err := client.ContainerStartWithResponse(context.TODO(), container)
	status_code := response.StatusCode()
	switch {
	case err != nil:
		fmt.Println("error requesting container start: ", err)

	case status_code == 200 && response.JSON200 == nil:
		fmt.Println("could not parse jocker engine response")

	case status_code == 200:
		return response.JSON200.Id

	case status_code == 304:
		fmt.Println("container already started")

	case status_code == 404:
		fmt.Println("no such container")

	case status_code == 500:
		fmt.Println("internal server error")

	default:
		fmt.Println("unknown status-code received from jocker engine: ", response.Status())
	}
	return ""
}

func NewContainerStopCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop one or more running containers",
		Run: func(cmd *cobra.Command, args []string) {
			RunContainerStop(cmd, args)
		},
	}
	return cmd
}

func RunContainerStop(cmd *cobra.Command, args []string) {
	name_or_id := args[0]
	client := NewHTTPClient()
	response, _ := client.ContainerStopWithResponse(context.TODO(), name_or_id)
	status_code := response.StatusCode()

	switch {
	case status_code == 204:
		fmt.Println("stopped ", response.JSON204.Id)

	case status_code == 304:
		fmt.Println("container already stopped")

	case status_code == 404:
		fmt.Println("no such container")

	case status_code == 500:
		fmt.Println("internal server error")

	default:
		fmt.Println("unknown status-code received from jocker engine: ", response.Status())
	}
}

func NewContainerListCommand() *cobra.Command {
	var all bool

	listCmd := &cobra.Command{
		Use:   "ls",
		Short: "List containers",
		Long:  `List containers loooong`,
		Run: func(cmd *cobra.Command, args []string) {
			RunContainerList(all, args)
		},
	}
	listCmd.Flags().BoolVarP(&all, "all", "a", false, "Show all containers (default shows just running)")
	return listCmd
}

func RunContainerList(all bool, args []string) {
	response, err := GetContainerList(all)
	if err == nil {
		PrintContainerList(response.JSON200)
	}
}

func GetContainerList(all bool) (*Openapi.ContainerListResponse, error) {
	params := Openapi.ContainerListParams{
		All: &all,
	}

	client := NewHTTPClient()

	response, err := client.ContainerListWithResponse(context.TODO(), &params)
	if err != nil {
		fmt.Println("Could not connect to jocker engine daemon: ", err)
		return response, err
	}

	if response.StatusCode() != 200 {
		fmt.Println("Jocker engine returned non-200 statuscode: ", response.Status())
		return response, errors.New("non-200 statuscode")
	}
	return response, nil
}

func PrintContainerList(container_list *[]Openapi.ContainerSummary) {
	fmt.Println(
		Cell("CONTAINER ID", 12), Sp(3),
		Cell("IMAGE", 15), Sp(3),
		Cell("COMMAND", 23), Sp(3),
		Cell("CREATED", 18), Sp(3),
		Cell("STATUS", 7), Sp(3),
		"NAME",
	)

	var running string

	for _, c := range *container_list {
		if *c.Running {
			running = "running"
		} else {
			running = "stopped"
		}
		created, _ := time.Parse(time.RFC3339, *c.Created)
		since_created := time.Since(created)

		fmt.Println(
			Cell(*c.Id, 12), Sp(1),
			Cell(*c.ImageId, 15), Sp(1),
			Cell(*c.Command, 23), Sp(1),
			Cell(HumanDuration(since_created)+" ago", 18), Sp(1),
			Cell(running, 7), Sp(1),
			*c.Name,
		)
	}
}

func NewHTTPClient() *Openapi.ClientWithResponses {
	client, err := Openapi.NewClientWithResponses(url)
	if err != nil {
		fmt.Println("Internal error: ", err)
		os.Exit(1)
	}
	return client
}

// HumanDuration returns a human-readable approximation of a duration
// (eg. "About a minute", "4 hours ago", etc.).
func HumanDuration(d time.Duration) string {
	if seconds := int(d.Seconds()); seconds < 1 {
		return "Less than a second"
	} else if seconds == 1 {
		return "1 second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else if minutes := int(d.Minutes()); minutes == 1 {
		return "About a minute"
	} else if minutes < 60 {
		return fmt.Sprintf("%d minutes", minutes)
	} else if hours := int(d.Hours() + 0.5); hours == 1 {
		return "About an hour"
	} else if hours < 48 {
		return fmt.Sprintf("%d hours", hours)
	} else if hours < 24*7*2 {
		return fmt.Sprintf("%d days", hours/24)
	} else if hours < 24*30*2 {
		return fmt.Sprintf("%d weeks", hours/24/7)
	} else if hours < 24*365*2 {
		return fmt.Sprintf("%d months", hours/24/30)
	}
	return fmt.Sprintf("%d years", int(d.Hours())/24/365)
}

func Cell(word string, max_len int) string {
	word_length := len(word)

	if word_length <= max_len {
		return word + Sp(max_len-word_length) + Sp(2)
	} else {
		return word[:max_len] + ".."
	}
}

func Sp(n int) string {
	return strings.Repeat(" ", n)
}
