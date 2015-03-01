package ralph

import (
	"log"
	"os"
	"os/exec"
	"reflect"
	"sync"
	"syscall"

	"github.com/bobrik/marathoner"
	"gopkg.in/yaml.v2"
)

type TwemproxyConfigurator struct {
	state twemproxyConfig
	mutex sync.Mutex
	conf  string
	bind  string
	args  []string
	pid   int
}

func NewTwemproxyConfigurator(conf string, bind string, args []string) *TwemproxyConfigurator {
	return &TwemproxyConfigurator{
		state: nil,
		mutex: sync.Mutex{},
		conf:  conf,
		bind:  bind,
		args:  args,
	}
}

func (t *TwemproxyConfigurator) Update(s marathoner.State, r *bool) error {
	err := t.update(s, r)
	if err != nil {
		log.Println("error updating configuration:", err)
	}

	return err
}

func (t *TwemproxyConfigurator) update(s marathoner.State, r *bool) error {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	log.Println("received update request")

	state := stateToPools(s, t.bind)

	if reflect.DeepEqual(state, t.state) {
		log.Println("state is the same, not doing any updates")
		*r = false
		return nil
	}

	t.state = state

	err := t.updateConfig()
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println("config updated")

	err = t.checkConfig()
	if err != nil {
		return err
	}

	log.Println("config validity checked")

	err = t.reload()
	if err != nil {
		return err
	}

	log.Println("twemproxy reloaded")

	*r = true
	return nil
}

func (t *TwemproxyConfigurator) updateConfig() error {
	c, err := yaml.Marshal(t.state)
	if err != nil {
		return err
	}

	f, err := os.Create(t.conf)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(c)

	return err
}

func (t *TwemproxyConfigurator) checkConfig() error {
	return exec.Command("nutcracker", "-t", "-c", t.conf).Run()
}

func (t *TwemproxyConfigurator) reload() error {
	if t.pid == 0 {
		log.Println("twemproxy is not started yet, starting")
		return t.runTwemproxy()
	}

	err := syscall.Kill(t.pid, syscall.SIGUSR1)
	if err != nil {
		if err != syscall.ESRCH {
			return err
		}

		return t.runTwemproxy()
	}

	return err
}

func (t *TwemproxyConfigurator) runTwemproxy() error {
	args := append([]string{"-c", "/etc/nutcracker.yml"}, t.args...)
	cmd := exec.Command("nutcracker", args...)
	err := cmd.Start()
	if err != nil {
		return err
	}

	t.pid = cmd.Process.Pid

	go cmd.Wait()

	return nil
}

func stateToPools(s marathoner.State, bind string) twemproxyConfig {
	c := map[string]twemproxyPool{}

	for n, a := range s {
		var name string

		if v, ok := a.Labels["twemproxy_pool"]; !ok {
			continue
		} else {
			name = v
		}

		pool, err := newPoolFromApp(a, bind)
		if err != nil {
			log.Printf("twemproxy creation error for %s: %s", n, err)
			continue
		}

		c[name] = pool
	}

	return c
}
