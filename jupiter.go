package copy

import (
	"context"
	"copy/pkg/flag"
	"copy/pkg/server"
	"copy/pkg/util/xdefer"
	"copy/pkg/worker"
	"copy/pkg/worker/xjob"
	"copy/pkg/xlog"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/douyu/jupiter"
	"github.com/douyu/jupiter/pkg/conf"
	"github.com/douyu/jupiter/pkg/ecode"
	"github.com/douyu/jupiter/pkg/registry"
	"github.com/douyu/jupiter/pkg/server/governor"
	"github.com/douyu/jupiter/pkg/util/xgo"
	xlog2 "github.com/douyu/jupiter/pkg/xlog"
	"github.com/prometheus/client_golang/prometheus"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	//StageAfterStop after app stop
	StageAfterStop uint32 = iota + 1
	//StageBeforeStop before app stop
	StageBeforeStop
)

// Application is the framework's instance, it contains the servers, workers, client and configuration settings.
// Create an instance of Application, by using &Application{}
type Application struct {
	cycle        *Cycle
	smu          *sync.RWMutex
	initOnce     sync.Once
	startupOnce  sync.Once
	stopOnce     sync.Once
	servers      []server.Server
	workers      []worker.Worker
	jobs         map[string]xjob.Runner
	logger       *xlog.Logger
	registerer   prometheus.Registry
	hooks        map[uint32]*xdefer.DeferStack
	configParser conf.Unmarshaller
	disableMap   map[jupiter.Disable]bool
}

func New(fns ...func() error) (*Application, error) {
	app := &Application{}
}

func DefaultApp() *Application {
	app := &Application{}
	return app
}

func (app *Application) intiHooks(hookKeys ...uint32) {
	app.hooks = make(map[uint32]*xdefer.DeferStack, len(hookKeys))
	for _, k := range hookKeys {
		app.hooks[k] = xdefer.NewStack()
	}
}

func (app *Application) runHooks(k uint32) {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Clean()
	}
}

func (app *Application) RegisterHooks(k uint32, fns ...func() error) error {
	hooks, ok := app.hooks[k]
	if ok {
		hooks.Push(fns...)
		return nil
	}
	return fmt.Errorf("hook stage not found")
}

func (app *Application) initialize() {
	app.initOnce.Do(func() {
		app.cycle = NewCycle()
		app.smu = &sync.RWMutex{}
		app.servers = make([]server.Server, 0)
		app.workers = make([]worker.Worker, 0)
		app.jobs = make(map[string]xjob.Runner)
		app.logger = JupiterLogger
		app.configParser = toml.Unmarshal
		app.disableMap = make(map[Disable]bool)

		app.initHooks(StageBeforeStop, StageAfterStop)
		app.SetRegistry(registry.Nop{})
	})
}

func (app *Application) startup() (err error) {
	app.startupOnce.Do(func() {
		err = xgo.SerialUntilError()
	})
	return
}

func (app *Application) Startup(fns ...func() error) error {
	app.initialize()
	if err := app.startup(); err != nil {
		return err
	}
	return xgo.SerialUntilError(fns...)()
}

func (app *Application) Serve(s ...server.Server) error {
	app.smu.Lock()
	defer app.smu.Unlock()
	app.servers = append(app.servers, s...)
	return nil
}

func (app *Application) Schedule(w worker.Worker) error {
	app.workers = append(app.workers, w)
	return nil
}

func (app *Application) Job(runner xjob.Runner) error {
	namedJob, ok := runner.(interface{ GetJobName() string })
	if !ok {
		return nil
	}

	jobName := namedJob.GetJobName()
	if flag.Bool("disable-job") {
		app.logger.Info("jupiter disable job", xlog.FieldName(jobName))
		return nil
	}

	// start job by name
	jobFlag := flag.String("job")
	if jobFlag == "" {
		app.logger.Error("jupiter jobs flag name empty", xlog.FieldName(jobName))
		return nil
	}

	if jobName != jobFlag {
		app.logger.Info("jupiter disable jobs", xlog.FieldName(jobName))
		return nil
	}
	app.logger.Info("jupiter register job", xlog.FieldName(jobName))
	app.jobs[jobName] = runner
	return nil
}

func (app *Application) SetRegistry(reg registry.Registry) {
	app.registerer = reg
}

func (app *Application) Run(servers ...server.Server) error {
	app.smu.Lock()
	app.servers = append(app.servers, servers...)
	app.smu.Unlock()

	app.waitSignals()
	defer app.Clean()

	app.startJobs()
	app.cycle.Run(app.s)

}

func (app *Application) clean() {
	_ = xlog2.DefaultLogger.Flush()
	_ = xlog.JupiterLogger.Flush()
}

func (app *Application) Stop() (err error) {
	app.stopOnce.Do(func() {
		app.runHooks(StageBeforeStop)

		if app.registerer != nil {
			err = app.registerer.Close()
			if err != nil {
				app.logger.Error("stop register close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
			}
		}
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(s.Stop)
			}(s)
		}
		app.smu.RUnlock()

		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(StageAfterStop)
		app.cycle.Close()
	})
	return
}

func (app *Application) GracefulStop(ctx context.Context)(err error) {
	app.stopOnce.Do(func() {
		app.runHooks(StageBeforeStop)

		if app.registerer != nil {
			err = app.registerer.Close()
			if err != nil {
				app.logger.Error("stop register close err", xlog.FieldMod(ecode.ModApp), xlog.FieldErr(err))
			}
		}
		//stop servers
		app.smu.RLock()
		for _, s := range app.servers {
			func(s server.Server) {
				app.cycle.Run(func() error {
					return s.GracefulStop(ctx)
				})
			}(s)
		}
		app.smu.RUnlock()

		//stop workers
		for _, w := range app.workers {
			func(w worker.Worker) {
				app.cycle.Run(w.Stop)
			}(w)
		}
		<-app.cycle.Done()
		app.runHooks(StageAfterStop)
		app.cycle.Close()
	})
	return err
}
func (app *Application) waitSignals() {
	app.logger.Info("init listen signal", xlog.FieldMod(ecode.ModApp), xlog.FieldEvent("init"))
	Shutdown(func(grace bool) {
		if grace {
			app.GracefulStop(context.TODO())
		} else {
			app.Stop()
		}
	})
}

func (app *Application)parseFlags()error  {
	if app.isDisable(jupiter.DisableParserFlag) {
		return nil
	}
	flag.Register(&flag.StringFlag{
		Name: "config",
		Usage: "--config",
		EnvVar: "JUPITER_CONFIG",
		Default: "",
		Action: func(name string, flagSet *flag.FlagSet) {

		},
	})
	flag.Register(&flag.BoolFlag{
		Name:    "watch",
		Usage:   "--watch, watch config change event",
		Default: false,
		EnvVar:  "JUPITER_CONFIG_WATCH",
	})

	flag.Register(&flag.BoolFlag{
		Name:    "version",
		Usage:   "--version, print version",
		Default: false,
		Action: func(string, *flag.FlagSet) {
			pkg.PrintVersion()
			os.Exit(0)
		},
	})
	return flag.Parse()

}

func (app *Application)loadConfig()error  {

}

var shutdownSignals = []os.Signal{syscall.SIGQUIT, os.Interrupt}

func Shutdown(stop func(grace bool)) {
	sig := make(chan os.Signal, 2)
	signal.Notify(
		sig,
		shutdownSignals, ...
	)
	go func() {
		s := <-sig
		go stop(s != syscall.SIGQUIT)
		<-sig
		os.Exit(128 + int(s.(syscall.Signal)))
	}()
}

func (app *Application) initGovernor() error {
	if app.isDisable(DisableDefaultGovernor) {
		app.logger.Info("defualt governor disable", xlog.FieldMod(ecode.ModApp))
		return nil
	}

	config := governor.StdConfig("governor")
	if !config.Enable {
		return nil
	}
}
