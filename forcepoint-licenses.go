package main

import (
	"fmt"
	"os"

	"github.com/Newlode/forcepoint-ngfw-licenses/codes"
	ngfwlicenses "github.com/Newlode/forcepoint-ngfw-licenses/ngfw-licenses"
	"github.com/logrusorgru/aurora"
	"github.com/mbndr/logo"
	"github.com/snwfdhmp/errlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	logger *logo.Logger

	cfg     config
	cfgFile string

	rootCmd = &cobra.Command{
		Use: "forcepoint-licenses",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			posList = ngfwlicenses.CreatePOSFormFiles()
		},
	}

	debug             bool
	verbose           bool
	concurrentWorkers int    = 8
	outputDir         string = "jar-files"

	posList ngfwlicenses.POSList
)

func init() {
	cobra.OnInitialize(initConfig)

	logger = logo.NewSimpleLogger(os.Stderr, logo.WARN, aurora.Cyan("MAIN  ").String(), true)
	ngfwlicenses.Logger = logo.NewSimpleLogger(os.Stderr, logo.WARN, aurora.Magenta("NGFWLIC").String(), true)

	// ConfigFile
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml in current directory)")

	// Debug / Verbose
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// ConcurrentWorkers
	rootCmd.PersistentFlags().IntVar(&concurrentWorkers, "concurrent-workers", concurrentWorkers, "Number of threads to use")
	viper.BindPFlag("concurrent_workers", rootCmd.PersistentFlags().Lookup("concurrent-workers"))

	// LicensesOutputDir
	rootCmd.PersistentFlags().StringVar(&outputDir, "output-dir", outputDir, "The directory where to store licenses files")
	viper.BindPFlag("licenses_output_dir", rootCmd.PersistentFlags().Lookup("output-dir"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.SetConfigFile("./config.yml")
	}

	readConfig()

	if verbose {
		logger.SetLevel(logo.INFO)
		ngfwlicenses.Logger.SetLevel(logo.INFO)
	}

	if debug {
		logger.SetLevel(logo.DEBUG)
		ngfwlicenses.Logger.SetLevel(logo.DEBUG)
	}
}

//=================================================================
// Commands

func runListCountries(cmd *cobra.Command, args []string) {
	markdown, _ := cmd.Flags().GetBool("markdown")
	if markdown {
		fmt.Printf(codes.CountriesToMarkdown())
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
		fmt.Printf(codes.StatesToMarkdown())
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
	posList.RefreshStatus(concurrentWorkers)
	posList.Display()
}

// runRegister
func runRegister(cmd *cobra.Command, args []string) {
	posList.RefreshStatus(concurrentWorkers)
	posList.Display()
	posList.Register(cfg.ConcurrentWorkers, cfg.ContactInfo, cfg.Reseller)
	posList.Display()
}

// runDownload
func runDownload(cmd *cobra.Command, args []string) {
	posList.RefreshStatus(concurrentWorkers)
	posList.Display()
	posList.Register(cfg.ConcurrentWorkers, cfg.ContactInfo, cfg.Reseller)
	posList.Display()
	posList.Download(cfg.ConcurrentWorkers, cfg.LicensesOutputDir)
}

// runDownloadOnly
func runDownloadOnly(cmd *cobra.Command, args []string) {
	_, err := os.Stat(cfg.LicensesOutputDir)

	if os.IsNotExist(err) {
		err = os.Mkdir(cfg.LicensesOutputDir, os.ModePerm)
		if errlog.Debug(err) {
			logger.Fatalf("Unable to create directory %s", cfg.LicensesOutputDir)
		}
	}
	posList.Download(cfg.ConcurrentWorkers, cfg.LicensesOutputDir)
}

// runNotImplemented
func runNotImplemented(cmd *cobra.Command, args []string) {
	logger.Fatalf("%s not yet implemented\n", cmd.Use)
}

//=================================================================
// Config

func readConfig() {
	// viper.SetConfigName("config.yml")
	viper.SetConfigType("yaml")
	// viper.AddConfigPath(".")

	//viper.SetDefault("concurrent_workers", 2)
	//viper.SetDefault("licenses_output_dir", "./out/")
	viper.SetDefault("contact_info", nil)

	viper.ReadInConfig()

	err := viper.Unmarshal(&cfg)
	if errlog.Debug(err) {
		logger.Fatalf("Unable to read config file: %s\n", err)
	}
}

type config struct {
	ConcurrentWorkers int    `mapstructure:"concurrent_workers"`
	LicensesOutputDir string `mapstructure:"licenses_output_dir"`

	ContactInfo *ngfwlicenses.ContactInfo `mapstructure:"contact_info"`
	Reseller    string                    `mapstructure:"resseller"`
}

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
		Args:             cobra.MaximumNArgs(1),
		ValidArgs:        codes.CountriesCodes,
		Run:              runListCountryStates,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	}
	cmdListCountryStates.Flags().Bool("markdown", false, "Use Markdown format output")

	var cmdVerify = &cobra.Command{
		Use:   "verify",
		Short: "Verify POS status",
		Args:  cobra.NoArgs,
		Run:   runVerify,
	}

	var cmdRegister = &cobra.Command{
		Use:    "register",
		Short:  "Verify and register all PoS",
		Args:   cobra.NoArgs,
		PreRun: runVerify,
		Run:    runRegister,
	}
	cmdRegister.Flags().StringArrayP("from-file", "f", nil, "filename")

	var cmdDownload = &cobra.Command{
		Use:     "download",
		Short:   "Verify, register and download licenses files for all PoS",
		Args:    cobra.NoArgs,
		Run:     runDownload,
		Aliases: []string{"register-and-download"},
	}

	var cmdDownloadOnly = &cobra.Command{
		Use:   "download-only",
		Short: "Verify and download licenses files for already registered PoS",
		Args:  cobra.NoArgs,
		Run:   runDownloadOnly,
	}

	var cmdInstall = &cobra.Command{
		Use:              "install",
		Short:            "Verify, register and download licenses on SMC for all PoS",
		Args:             cobra.NoArgs,
		Run:              runNotImplemented,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	}

	var cmdInstallOnly = &cobra.Command{
		Use:              "install-only",
		Short:            "Verify and install licenses on SMC for already registered PoS",
		Args:             cobra.NoArgs,
		Run:              runNotImplemented,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {},
	}

	rootCmd.AddCommand(
		cmdListCountries, cmdListCountryStates,
		cmdVerify,
		cmdRegister,
		cmdDownload, cmdDownloadOnly,
		cmdInstall, cmdInstallOnly)
	rootCmd.Execute()
}
