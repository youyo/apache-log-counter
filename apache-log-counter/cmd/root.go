package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"

	"github.com/Songmu/axslogparser"
	"github.com/fatih/color"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/youyo/apache-log-counter"
)

var (
	Filter  string
	LogFile string
	Count   string
	Debug   bool
)

func init() {}

func NewCmdRoot() *cobra.Command {
	cmd := &cobra.Command{
		Use:   `apache-log-counter -l /path/to/logfile -c host`,
		Short: "",
		Run: func(cmd *cobra.Command, args []string) {
			// initialize
			alc := apache_log_counter.NewApacheLogCounter()

			// parse json params
			if err := alc.ParseFilter(Filter); err != nil {
				cmd.SetOutput(os.Stderr)
				err = errors.Wrap(err,
					"failed to parse filter parameter. ",
				)
				cmd.Println(err)
				os.Exit(1)
			}

			// load log file
			fp, err := os.Open(LogFile)
			if err != nil {
				cmd.SetOutput(os.Stderr)
				err = errors.Wrap(err,
					"can't load log file. must be required log file path. ",
				)
				cmd.Println(err)
				os.Exit(1)
			}
			defer fp.Close()

			scanner := bufio.NewScanner(fp)

			Counter := make(map[string]int)
			for scanner.Scan() {
				log, err := axslogparser.Parse(scanner.Text())
				if err != nil {
					if Debug {
						cmd.SetOutput(os.Stderr)
						err = errors.Wrap(err,
							"parse apache log file error. ",
						)
						color.Set(color.FgRed)
						cmd.Printf("%s\n\n", err)
						color.Unset()
					}
					continue
				}

				// filtering logs
				// host
				if alc.FilteringHost(log.Host) {
					continue
				}

				// remote host
				if alc.FilteringRemoteHost(log.VirtualHost) {
					continue
				}

				// status
				if alc.FilteringStatus(log.Status) {
					continue
				}

				// method
				if alc.FilteringMethod(log.Method) {
					continue
				}

				// request_uri
				if alc.FilteringRequestURI(log.RequestURI) {
					continue
				}

				// request
				if alc.FilteringRequest(log.Request) {
					continue
				}

				// start_time
				if b, err := alc.FilteringStartTime(log.Time); err != nil {
					cmd.SetOutput(os.Stderr)
					err = errors.Wrap(err,
						"failed to parse start_time. ",
					)
					cmd.Println(err)
					continue
				} else if b {
					continue
				}

				// end_time
				if b, err := alc.FilteringEndTime(log.Time); err != nil {
					cmd.SetOutput(os.Stderr)
					err = errors.Wrap(err,
						"failed to parse end_time. ",
					)
					cmd.Println(err)
					continue
				} else if b {
					continue
				}

				// aggregate
				switch Count {
				case "host":
					Counter[log.Host]++
				case "remote_host":
					Counter[log.VirtualHost]++
				case "status":
					Counter[strconv.Itoa(log.Status)]++
				case "request_uri":
					Counter[log.RequestURI]++
				case "request":
					Counter[log.Request]++
				case "method":
					Counter[log.Method]++
				default:
					cmd.SetOutput(os.Stderr)
					err := errors.New("count parameter is invalid.")
					cmd.Println(err)
					os.Exit(1)
				}
			}

			// sort
			c := apache_log_counter.SortDesc(Counter)

			// output
			wtr := bufio.NewWriter(os.Stdout)
			for k, v := range c {
				if Count == "remote_host" {
					isoCode, country, err := alc.FetchRemoteHostCountry(v.Key)
					if err != nil {
						if Debug {
							cmd.SetOutput(os.Stderr)
							err = errors.Wrap(err,
								"parse apache log file error. ",
							)
							color.Set(color.FgRed)
							cmd.Printf("%s\n\n", err)
							color.Unset()
						}
					}
					fmt.Fprintln(wtr, v.Key, isoCode, country, v.Value)
				} else {
					fmt.Fprintln(wtr, v.Key, v.Value)
				}
				if k > 10 {
					break
				}
			}
			wtr.Flush()

			// if error happened
			if err = scanner.Err(); err != nil {
				cmd.SetOutput(os.Stderr)
				err = errors.Wrap(err,
					"occured scanner errors. . ",
				)
				cmd.Println(err)
			}
		},
	}

	cobra.OnInitialize(initConfig)

	cmd.Flags().StringVarP(&Filter, "filter", "f", "", `filtering conditions by json syntax. (optional)
example)
	-f '{
		"host":"example.com",
		"remote_host":"some_ipaddress_xxx.xxx.xxx.xxx",
		"status":200,
		"request_uri":"/xmlrpc.php",
		"request":"POST /xmlrpc.php HTTP/1.1",
		"method":"POST"
	}'`)
	cmd.Flags().StringVarP(&LogFile, "logfile", "l", "", "path to log file. (required)")
	cmd.Flags().StringVarP(&Count, "count", "c", "", `count by parameter. (required)
[
	'host',
	'remote_host',
	'status',
	'request_uri',
	'request',
	'method'
]`)
	cmd.Flags().BoolVarP(&Debug, "debug", "d", false, "display debug info. (optional)")

	return cmd
}
