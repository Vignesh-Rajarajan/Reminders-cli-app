package client

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

type idsFlag []string

func (list idsFlag) String() string {
	return strings.Join(list, ",")
}

func (list *idsFlag) Set(v string) error {
	*list = append(*list, v)
	return nil
}

type BackendHTTPClient interface {
	Create(title, message string, duration time.Duration) ([]byte, error)
	Edit(id, title, message string, duration time.Duration) ([]byte, error)
	Fetch(ids []string) ([]byte, error)
	Delete(ids []string) error
	Healthy(host string) bool
}

type Switch struct {
	client        BackendHTTPClient
	backendAPIURL string
	commands      map[string]func() func(string) error
}

func NewSwitch(uri string) Switch {
	httpClient := NewHTTPClient(uri)
	s := Switch{
		client:        httpClient,
		backendAPIURL: uri,
	}
	s.commands = map[string]func() func(string) error{
		"create": s.create,
		"edit":   s.edit,
		"fetch":  s.fetch,
		"delete": s.delete,
		"health": s.health,
	}
	return s
}

func (s Switch) Help() {
	var help string

	for name := range s.commands {
		help += name + "\t --help \n"
	}
	fmt.Printf("Usage of: %s\n <command> [args] \n %s", os.Args[0], help)
}
func (s Switch) Switch() error {
	cmdName := os.Args[1]

	cmd, ok := s.commands[cmdName]

	if !ok {
		return fmt.Errorf("invalid command passed '%s' ", cmdName)
	}
	return cmd()(cmdName)
}

func (s Switch) create() func(string) error {
	return func(cmd string) error {
		createCmd := flag.NewFlagSet(cmd, flag.ExitOnError)
		t, m, d := s.reminderFlags(createCmd)

		if err := s.checkArgs(3); err != nil {
			return err
		}
		if err := s.parseCmd(createCmd); err != nil {
			return err
		}

		res, err := s.client.Create(*t, *m, *d)

		if err != nil {
			wrapError("Could not create reminder", err)
		}

		fmt.Printf("created reminder successfully : \n%s", string(res))
		return nil
	}
}

func (s Switch) edit() func(string) error {
	return func(cmd string) error {

		ids := idsFlag{}
		editCmd := flag.NewFlagSet(cmd, flag.ExitOnError)

		editCmd.Var(&ids, "id", "The ID(int) of the reminder to edit")
		t, m, d := s.reminderFlags(editCmd)
		if err := s.checkArgs(2); err != nil {
			return err
		}
		if err := s.parseCmd(editCmd); err != nil {
			return err
		}
		lastId := ids[len(ids)-1]
		res, err := s.client.Edit(lastId, *t, *m, *d)

		if err != nil {
			wrapError("Could not Edit the reminder", err)
		}

		fmt.Printf("Edited the reminder successfully : \n%s", string(res))

		return nil
	}
}

func (s Switch) fetch() func(string) error {
	return func(cmd string) error {

		ids := idsFlag{}
		fetchCmd := flag.NewFlagSet(cmd, flag.ExitOnError)

		fetchCmd.Var(&ids, "id", "List of reminders ID(s) to fetch")

		if err := s.checkArgs(1); err != nil {
			return err
		}
		if err := s.parseCmd(fetchCmd); err != nil {
			return err
		}

		res, err := s.client.Fetch(ids)

		if err != nil {
			wrapError("Could not Fetch the reminder(s)", err)
		}

		fmt.Printf("Fetched the reminder(s) successfully : \n%s", string(res))

		return nil
	}
}

func (s Switch) delete() func(string) error {
	return func(cmd string) error {

		ids := idsFlag{}
		deleteCmd := flag.NewFlagSet(cmd, flag.ExitOnError)

		deleteCmd.Var(&ids, "id", "List of reminders ID(s) to delete")

		if err := s.checkArgs(1); err != nil {
			return err
		}
		if err := s.parseCmd(deleteCmd); err != nil {
			return err
		}

		err := s.client.Delete(ids)

		if err != nil {
			wrapError("Could not Delete the reminder(s)", err)
		}

		fmt.Printf("Deleted the reminder(s) successfully : \n%v\n", ids)

		return nil
	}
}

func (s Switch) health() func(string) error {
	return func(cmd string) error {
		var host string
		healthCmd := flag.NewFlagSet(cmd, flag.ExitOnError)
		healthCmd.StringVar(&host, "host", s.backendAPIURL, "Host to ping for health")
		if err := s.parseCmd(healthCmd); err != nil {
			return err
		}
		if !s.client.Healthy(host) {
			fmt.Printf("the host %s is down", host)
		} else {
			fmt.Printf("the host %s is up and running", host)
		}
		return nil
	}

}

func (s Switch) reminderFlags(f *flag.FlagSet) (*string, *string, *time.Duration) {
	t, m, d := "", "", time.Duration(0)
	f.StringVar(&t, "title", "", "Reminder title")
	f.StringVar(&t, "t", "", "Reminder title")
	f.StringVar(&m, "message", "", "Reminder message")
	f.StringVar(&m, "m", "", "Reminder message")
	f.DurationVar(&d, "duration", time.Duration(0), "Reminder time")
	f.DurationVar(&d, "d", time.Duration(0), "Reminder time")

	return &t, &m, &d
}

func (s Switch) parseCmd(cmd *flag.FlagSet) error {
	err := cmd.Parse(os.Args[:2])
	if err != nil {
		return wrapError("could not parse '"+cmd.Name()+"' command flags", err)
	}
	return nil
}

func (s Switch) checkArgs(minArgs int) error {
	if len(os.Args) == 3 && os.Args[2] == "--help" {
		return nil
	}

	if len(os.Args)-2 < minArgs {
		fmt.Printf(
			"incorrect use of %s\n%s %s --help\n",
			os.Args[1], os.Args[0], os.Args[1],
		)

		return fmt.Errorf(
			"%s expects at least %d arg(s), provided %d args",
			os.Args[1], minArgs, len(os.Args)-2)
	}

	return nil
}
