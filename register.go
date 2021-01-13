package main

import (
	"fmt"
	"os"

	"github.com/logrusorgru/aurora"
	"github.com/mbndr/logo"
	"github.com/snwfdhmp/errlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	codes "gitlab.com/newlodegroup/forcepoint/ngfw-licences.git/codes"
	ngfwlicences "gitlab.com/newlodegroup/forcepoint/ngfw-licences.git/ngfw-licences"
)

var (
	logger *logo.Logger

	cfg     config
	cfgFile string

	rootCmd = &cobra.Command{
		Use: "register",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			posList = ngfwlicences.CreatePOSFormFiles()
		},
	}

	debug             bool
	verbose           bool
	concurrentWorkers int

	posList ngfwlicences.POSList
)

func init() {
	cobra.OnInitialize(initConfig)

	logger = logo.NewSimpleLogger(os.Stderr, logo.WARN, aurora.Cyan("MAIN  ").String(), true)
	ngfwlicences.Logger = logo.NewSimpleLogger(os.Stderr, logo.WARN, aurora.Magenta("NGFWLIC").String(), true)

	// ConfigFile
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is config.yml in current directory)")

	// Debug / Verbose
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")

	// ConcurrentWorkers
	rootCmd.PersistentFlags().IntVar(&concurrentWorkers, "concurrent-workers", 8, "Number of threads to use")
	viper.BindPFlag("concurrent_workers", rootCmd.PersistentFlags().Lookup("concurrent-workers"))
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
		ngfwlicences.Logger.SetLevel(logo.INFO)
	}

	if debug {
		logger.SetLevel(logo.DEBUG)
		ngfwlicences.Logger.SetLevel(logo.DEBUG)
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
	posList.CheckValidity(concurrentWorkers)

	posOk := posList.GetByNotStatus(ngfwlicences.Invalid)
	fmt.Printf("Found %d valid PoS:\n", len(posOk))
	for _, pos := range posOk {
		fmt.Printf("- %v\n", pos.DetailedString())
	}

	posKo := posList.GetByStatus(ngfwlicences.Invalid)
	if len(posKo) > 0 {
		fmt.Printf("\nFound %d invalid PoS:\n", len(posKo))
		for _, pos := range posKo {
			fmt.Printf("- %v\n", pos.DetailedString())
		}
	}

}

// runRegister
func runRegister(cmd *cobra.Command, args []string) {

}

func runGenCountriesMD(cmd *cobra.Command, args []string) {
	fmt.Println(codes.CountriesToMarkdown())
}

//=================================================================
// Config

func readConfig() {
	// viper.SetConfigName("config.yml")
	viper.SetConfigType("yaml")
	// viper.AddConfigPath(".")

	viper.SetDefault("concurrent_workers", 2)
	viper.SetDefault("licences_output_dir", "./out/")
	viper.SetDefault("contact_info", nil)

	err := viper.ReadInConfig()
	/*
		if errlog.Debug(err) {
			//	logger.Fatalf("Fatal error while opening config file: %s\n", err)
			return
		}
	*/

	err = viper.Unmarshal(&cfg)
	if errlog.Debug(err) {
		logger.Fatalf("Unable to read config file: %s\n", err)
	}
}

type config struct {
	ConcurrentWorkers int    `mapstructure:"concurrent_workers"`
	LicencesOutputDir string `mapstructure:"licences_output_dir"`

	ContactInfo *struct {
		Firstname string `mapstructure:"firstname"`
		Lastname  string `mapstructure:"lastname"`
		Email     string `mapstructure:"email"`
		Phone     string `mapstructure:"phone"`
		Company   string `mapstructure:"company*"`
		Address   string `mapstructure:"address"`
		Zip       string `mapstructure:"zip"`
		City      string `mapstructure:"city"`
		Country   string `mapstructure:"country"`
		State     string `mapstructure:"state"`
	} `mapstructure:"contact_info"`
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
		Short: "Check PoS validity",
		Long:  `Check the validity of all PoS contained in HTML files.`,
		Args:  cobra.NoArgs,
		Run:   runVerify,
	}

	var cmdRegister = &cobra.Command{
		Use:    "register",
		Short:  "Register all PoS",
		Long:   `Register all PoS found in HTML files.`,
		Args:   cobra.NoArgs,
		PreRun: runVerify,
		Run:    runRegister,
	}

	var cmdDownload = &cobra.Command{
		Use:   "download",
		Short: "Verify, register and download licence files for all PoS",
		Long: `echo things multiple times back to the user by providing
a count and a string.`,
		Args: cobra.NoArgs,
		Run:  runVerify,
	}

	var cmdDownloadOnly = &cobra.Command{
		Use:   "download-only",
		Short: "Verify and download licence files for already registered PoS",
		Long: `echo things multiple times back to the user by providing
a count and a string.`,
		Args: cobra.NoArgs,
		Run:  runVerify,
	}

	var cmdInstall = &cobra.Command{
		Use:   "install",
		Short: "Echo anything to the screen more times",
		Long: `echo things multiple times back to the user by providing
a count and a string.`,
		Args: cobra.NoArgs,
		Run:  runVerify,
	}

	rootCmd.AddCommand(cmdListCountries, cmdListCountryStates, cmdVerify, cmdRegister, cmdDownload, cmdDownloadOnly, cmdInstall)
	rootCmd.Execute()
}
