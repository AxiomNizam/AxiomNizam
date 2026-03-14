package main

func main() {
	if err := RootCmd.Execute(); err != nil {
		handleCommandError(err)
	}
}
