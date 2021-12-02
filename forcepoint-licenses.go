package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Newlode/forcepoint-ngfw-licenses/codes"
	"github.com/Newlode/forcepoint-ngfw-licenses/config"
	ngfwlicenses "github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses"
	"github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses/pox"
	"github.com/logrusorgru/aurora"
	"github.com/mbndr/logo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger *logo.Logger

	cfg = &config.Cfg

	rootCmd = &cobra.Command{
		Use:   "forcepoint-licenses",
		Short: "forcepoint-licenses is a tool help you to register Forcepoint Next Generation Firewalls licenses",
		Long: `forcepoint-licenses is a tool to verify, register and download licenses for Forcepoint Next Generation Firewalls
		
Please report issues or feature requests using Github project page at https://github.com/Newlode/forcepoint-ngfw-licenses/issues
		
Written by Newlode https://www.newlode.io
`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			format, _ := cmd.Flags().GetString("format")
			debug, _ := cmd.Flags().GetBool("debug")
			cfg.Silent = format == "json" || format == "csv" || debug

			if posOnly && polOnly {
				logger.Fatalf("--pos-only and --pol-only are mutually exclusive")
			}

			poxList = pox.ReadPoXFormArgs(args, polOnly, posOnly)
		},
	}

	poxList pox.PoXList

	verifyFormat string
	posOnly      bool
	polOnly      bool
)

func init() {
	logger = config.GetNewLogger(aurora.Cyan("MAIN  ").String())
	ngfwlicenses.Logger = config.GetNewLogger(aurora.Magenta("NGFWLIC").String())
	pox.Logger = config.GetNewLogger(aurora.Green("POX   ").String())
	config.Logger = config.GetNewLogger(aurora.Red("CONFIG").String())

	// ConfigFile
	rootCmd.PersistentFlags().StringVarP(&config.ConfigFile, "config", "c", "", "config file (default is config.yml in current directory)")

	// PoSOnly / PoLOnly
	rootCmd.PersistentFlags().BoolVar(&posOnly, "pos-only", false, "PoS only")
	rootCmd.PersistentFlags().BoolVar(&polOnly, "pol-only", false, "PoL-only")

	// Debug / Verbose
	rootCmd.PersistentFlags().BoolVarP(&cfg.Debug, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&cfg.Verbose, "verbose", "v", false, "Enable verbose output")

	// ConcurrentWorkers
	rootCmd.PersistentFlags().IntVar(&cfg.ConcurrentWorkers, "concurrent-workers", 8, "Number of threads to use")
	viper.BindPFlag("concurrent_workers", rootCmd.PersistentFlags().Lookup("concurrent-workers"))

	// LicensesOutputDir
	rootCmd.PersistentFlags().StringVar(&cfg.LicensesOutputDir, "output-dir", "jar-files", "The directory where to store licenses files")
	viper.BindPFlag("licenses_output_dir", rootCmd.PersistentFlags().Lookup("output-dir"))

}

//=================================================================
// Commands

func runListCountries(cmd *cobra.Command, args []string) {
	markdown, _ := cmd.Flags().GetBool("markdown")
	if markdown {
		fmt.Println(codes.CountriesToMarkdown())
		return
	}
	for _, code := range codes.CountriesCodes {
		if cfg.ContactInfo != nil && cfg.ContactInfo.Country == code {
			fmt.Printf("\n%s: %s (selected)\n\n", aurora.Green(code), aurora.Green(codes.Countries[code]))
		} else {
			fmt.Printf("%s: %s\n", code, aurora.Gray(12, codes.Countries[code]))
		}
	}
}

func runListCountryStates(cmd *cobra.Command, args []string) {
	markdown, _ := cmd.Flags().GetBool("markdown")
	if markdown {
		fmt.Println(codes.StatesToMarkdown())
		// codes.StatesToMarkdown()
		return
	}
	countryCode := args[0]
	for _, code := range codes.StatesCodes[args[0]] {
		if cfg.ContactInfo != nil && cfg.ContactInfo.Country == countryCode && cfg.ContactInfo.State == code {
			fmt.Printf("\n%s: %s (selected)\n\n", aurora.Green(code), aurora.Green(codes.States[countryCode][code]))
		} else {
			fmt.Printf("%s: %s\n", code, aurora.Gray(12, codes.States[countryCode][code]))
		}
	}
}

// runVerify just check online the PoS status
func runVerify(cmd *cobra.Command, args []string) {
	poxList.RefreshStatus()
	switch verifyFormat {
	case "json":
		obj := struct {
			PoLList pox.PoXList `json:"pol_list,omitempty"`
			PoSList pox.PoXList `json:"pos_list,omitempty"`
		}{
			PoLList: poxList.GetAllPoL(),
			PoSList: poxList.GetAllPoS(),
		}
		out, _ := json.MarshalIndent(obj, "", "  ")
		fmt.Println(string(out))
	case "csv":
		w := csv.NewWriter(os.Stdout)
		w.Write([]string{"PoS", "PoL", "LicenseStatus", "LicenseID", "ProductName", "Binding", "Platform", "LicensePeriod", "SerialNumber", "MaintenanceStatus", "MaintenanceEndDate", "Company"})
		for _, r := range poxList {
			line := []string{r.PoS, r.PoL, string(r.Status), r.LicenseID, r.ProductName, r.Binding, r.Platform, r.LicensePeriod, r.SerialNumber, string(r.MaintenanceStatus), r.MaintenanceEndDate, r.Company}
			if err := w.Write(line); err != nil {
				log.Fatalln("error writing record to csv:", err)
			}
		}
		w.Flush()
	default:
		poxList.Display()
	}
}

// runRegister
func runRegister(cmd *cobra.Command, args []string) {
	poxList.RefreshStatus()
	poxList.Display()
	poxList.Register()
	poxList.Display()
}

// runDownload
func runDownload(cmd *cobra.Command, args []string) {
	poxList.RefreshStatus()
	poxList.Display()
	poxList.Register()
	poxList.Download()
}

// runDownloadOnly
func runDownloadOnly(cmd *cobra.Command, args []string) {
	poxList.RefreshStatus()
	poxList.Display()
	poxList.Download()
}

// runChangeBinding
func runChangeBinding(cmd *cobra.Command, args []string) {
	poxList.RefreshStatus()
	poxList.ChangeBinding()
	// poxList.RefreshStatus()
	poxList.Display()
}

// runNotImplemented
/*
func runNotImplemented(cmd *cobra.Command, args []string) {
	logger.Fatalf("%s not yet implemented\n", cmd.Use)
}
*/

//=================================================================
// main()

func main() {

	var cmdListCountries = &cobra.Command{
		Use:              "list-countries",
		Short:            "Display the list of countries and their codes",
		Args:             cobra.NoArgs,
		Run:              runListCountries,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	}
	cmdListCountries.Flags().Bool("markdown", false, "Use Markdown format output")

	var cmdListCountryStates = &cobra.Command{
		Use:              "list-country-states [country-code]",
		Short:            "Display the list of the states of a country and their codes",
		Args:             cobra.ExactArgs(1),
		ValidArgs:        codes.CountriesCodes,
		Run:              runListCountryStates,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	}
	cmdListCountryStates.Flags().Bool("markdown", false, "Use Markdown format output")

	var cmdVerify = &cobra.Command{
		Use:   "verify",
		Short: "Verify PoS status",
		Args:  cobra.ArbitraryArgs,
		Run:   runVerify,
	}
	cmdVerify.Flags().StringVarP(&verifyFormat, "format", "f", "none", "Choose a specific output format [none|csv|json]")

	var cmdRegister = &cobra.Command{
		Use:    "register",
		Short:  "Verify and register all PoS",
		Args:   cobra.ArbitraryArgs,
		PreRun: runVerify,
		Run:    runRegister,
	}
	//? cmdRegister.Flags().StringArrayP("from-file", "f", nil, "filename")

	var cmdDownload = &cobra.Command{
		Use:     "download",
		Short:   "Verify, register and download licenses files for all PoS",
		Args:    cobra.ArbitraryArgs,
		Run:     runDownload,
		Aliases: []string{"register-and-download"},
	}

	var cmdDownloadOnly = &cobra.Command{
		Use:   "download-only",
		Short: "Verify and download licenses files for already registered PoS",
		Args:  cobra.ArbitraryArgs,
		Run:   runDownloadOnly,
	}

	var cmdChangeBinding = &cobra.Command{
		Use:   "change-binding",
		Short: "Change binding already registered PoS",
		Args:  cobra.ArbitraryArgs,
		Run:   runChangeBinding,
	}

	/*
		var cmdInstall = &cobra.Command{
			Use:              "install",
			Short:            "Verify, register and download licenses on SMC for all PoS",
			Args:             cobra.NoArgs,
			Run:              runNotImplemented,
			PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		}
	*/

	/*
		var cmdInstallOnly = &cobra.Command{
			Use:              "install-only",
			Short:            "Verify and install licenses on SMC for already registered PoS",
			Args:             cobra.NoArgs,
			Run:              runNotImplemented,
			PersistentPreRun: func(cmd *cobra.Command, args []string) {},
		}
	*/

	rootCmd.AddCommand(
		cmdListCountries, cmdListCountryStates,
		cmdVerify,
		cmdRegister,
		cmdDownload, cmdDownloadOnly,
		cmdChangeBinding,
		//* cmdInstall, cmdInstallOnly,
	)
	rootCmd.Execute()
}
