package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"tacacs-test-client/tacacs"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	TACACSServer string `mapstructure:"tacacs_server"`
	TACACSSecret string `mapstructure:"tacacs_secret"`
	TestUsername string `mapstructure:"test_username"`
	TestPassword string `mapstructure:"test_password"`
	LogLevel     string `mapstructure:"log_level"`
}

type TestCase struct {
	Name        string
	Username    string
	Password    string
	Command     string
	ExpectedAuth bool
	ExpectedAuthz bool
}

func main() {
	var runTests = flag.Bool("test", false, "Run automated tests")
	var interactive = flag.Bool("interactive", false, "Run in interactive mode")
	var username = flag.String("username", "", "Username for authentication")
	var password = flag.String("password", "", "Password for authentication")
	var command = flag.String("command", "", "Command to authorize")
	flag.Parse()

	cfg := loadConfig()
	
	logger := logrus.New()
	level, err := logrus.ParseLevel(cfg.LogLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	if *runTests {
		runAutomatedTests(cfg, logger)
	} else if *interactive {
		runInteractiveMode(cfg, logger)
	} else if *username != "" && *password != "" {
		testSingleUser(cfg, logger, *username, *password, *command)
	} else {
		fmt.Println("Usage:")
		fmt.Println("  -test           Run automated test suite")
		fmt.Println("  -interactive    Run in interactive mode")
		fmt.Println("  -username       Username for single test")
		fmt.Println("  -password       Password for single test")
		fmt.Println("  -command        Command to test authorization")
		os.Exit(1)
	}
}

func loadConfig() *Config {
	viper.SetDefault("tacacs_server", "localhost:49")
	viper.SetDefault("tacacs_secret", "testing123")
	viper.SetDefault("test_username", "testuser")
	viper.SetDefault("test_password", "testpass")
	viper.SetDefault("log_level", "info")

	viper.AutomaticEnv()

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		logrus.WithError(err).Fatal("Failed to unmarshal configuration")
	}

	return &config
}

func runAutomatedTests(cfg *Config, logger *logrus.Logger) {
	logger.Info("Starting automated TACACS+ tests")

	testCases := []TestCase{
		{
			Name:         "Admin Authentication",
			Username:     "admin",
			Password:     "admin123",
			Command:      "configure terminal",
			ExpectedAuth: true,
			ExpectedAuthz: true,
		},
		{
			Name:         "User Authentication",
			Username:     "netuser",
			Password:     "user123",
			Command:      "show running-config",
			ExpectedAuth: true,
			ExpectedAuthz: true,
		},
		{
			Name:         "ReadOnly Authentication", 
			Username:     "readonly",
			Password:     "readonly123",
			Command:      "show version",
			ExpectedAuth: true,
			ExpectedAuthz: true,
		},
		{
			Name:         "ReadOnly Denied Command",
			Username:     "readonly",
			Password:     "readonly123",
			Command:      "configure terminal",
			ExpectedAuth: true,
			ExpectedAuthz: false,
		},
		{
			Name:         "Invalid Credentials",
			Username:     "invalid",
			Password:     "invalid",
			Command:      "show version",
			ExpectedAuth: false,
			ExpectedAuthz: false,
		},
		{
			Name:         "Test User",
			Username:     cfg.TestUsername,
			Password:     cfg.TestPassword,
			Command:      "show interfaces",
			ExpectedAuth: true,
			ExpectedAuthz: true,
		},
	}

	passed := 0
	failed := 0

	for _, tc := range testCases {
		logger.WithField("test", tc.Name).Info("Running test case")
		
		if runTestCase(cfg, logger, tc) {
			logger.WithField("test", tc.Name).Info("✓ PASSED")
			passed++
		} else {
			logger.WithField("test", tc.Name).Error("✗ FAILED")
			failed++
		}
		
		time.Sleep(1 * time.Second)
	}

	logger.WithFields(logrus.Fields{
		"passed": passed,
		"failed": failed,
		"total":  len(testCases),
	}).Info("Test results")

	if failed > 0 {
		os.Exit(1)
	}
}

func runTestCase(cfg *Config, logger *logrus.Logger, tc TestCase) bool {
	success := true

	authResult := testAuthentication(cfg, logger, tc.Username, tc.Password)
	if authResult != tc.ExpectedAuth {
		logger.WithFields(logrus.Fields{
			"expected": tc.ExpectedAuth,
			"actual":   authResult,
		}).Error("Authentication test failed")
		success = false
	}

	if tc.ExpectedAuth && authResult {
		authzResult := testAuthorization(cfg, logger, tc.Username, tc.Command)
		if authzResult != tc.ExpectedAuthz {
			logger.WithFields(logrus.Fields{
				"expected": tc.ExpectedAuthz,
				"actual":   authzResult,
			}).Error("Authorization test failed")
			success = false
		}

		testAccounting(cfg, logger, tc.Username, tc.Command)
	}

	return success
}

func testSingleUser(cfg *Config, logger *logrus.Logger, username, password, command string) {
	logger.WithFields(logrus.Fields{
		"username": username,
		"command":  command,
	}).Info("Testing single user")

	if testAuthentication(cfg, logger, username, password) {
		logger.Info("✓ Authentication successful")
		
		if command != "" {
			if testAuthorization(cfg, logger, username, command) {
				logger.Info("✓ Authorization successful")
			} else {
				logger.Error("✗ Authorization failed")
			}
			
			testAccounting(cfg, logger, username, command)
		}
	} else {
		logger.Error("✗ Authentication failed")
	}
}

func runInteractiveMode(cfg *Config, logger *logrus.Logger) {
	logger.Info("Starting interactive TACACS+ client")
	
	for {
		fmt.Print("Enter username (or 'quit' to exit): ")
		var username string
		fmt.Scanln(&username)
		
		if username == "quit" {
			break
		}
		
		fmt.Print("Enter password: ")
		var password string
		fmt.Scanln(&password)
		
		if testAuthentication(cfg, logger, username, password) {
			fmt.Println("✓ Authentication successful")
			
			for {
				fmt.Print("Enter command (or 'logout' to switch user): ")
				var command string
				fmt.Scanln(&command)
				
				if command == "logout" {
					break
				}
				
				if testAuthorization(cfg, logger, username, command) {
					fmt.Println("✓ Command authorized")
				} else {
					fmt.Println("✗ Command denied")
				}
				
				testAccounting(cfg, logger, username, command)
			}
		} else {
			fmt.Println("✗ Authentication failed")
		}
	}
}

func testAuthentication(cfg *Config, logger *logrus.Logger, username, password string) bool {
	conn, err := net.Dial("tcp", cfg.TACACSServer)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to TACACS+ server")
		return false
	}
	defer conn.Close()

	client := tacacs.NewClient(conn, []byte(cfg.TACACSSecret))

	success, err := client.Authenticate(username, password)
	if err != nil {
		logger.WithError(err).Debug("Authentication request failed")
		return false
	}

	return success
}

func testAuthorization(cfg *Config, logger *logrus.Logger, username, command string) bool {
	conn, err := net.Dial("tcp", cfg.TACACSServer)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to TACACS+ server")
		return false
	}
	defer conn.Close()

	client := tacacs.NewClient(conn, []byte(cfg.TACACSSecret))

	success, err := client.Authorize(username, command)
	if err != nil {
		logger.WithError(err).Debug("Authorization request failed")
		return false
	}

	return success
}

func testAccounting(cfg *Config, logger *logrus.Logger, username, command string) {
	conn, err := net.Dial("tcp", cfg.TACACSServer)
	if err != nil {
		logger.WithError(err).Error("Failed to connect to TACACS+ server")
		return
	}
	defer conn.Close()

	client := tacacs.NewClient(conn, []byte(cfg.TACACSSecret))

	err = client.Account(username, command)
	if err != nil {
		logger.WithError(err).Debug("Accounting request failed")
	}
}