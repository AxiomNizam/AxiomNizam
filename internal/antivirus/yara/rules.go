package yara

// ─────────────────────────────────────────────────────────────────────────────
// Built-in YARA Rules
//
// High-value detection rules compiled into the binary for baseline
// coverage without any external rule files. Categories:
//
//   - Ransomware (file encryption markers, ransom notes)
//   - Cryptominer (mining pool URLs, miner binaries)
//   - Webshell (PHP/JSP/ASPX backdoors)
//   - Exploit kits (common payloads)
//   - Dropper / loader patterns
// ─────────────────────────────────────────────────────────────────────────────

// builtinRuleSources contains YARA rules as raw strings.
var builtinRuleSources = []string{

	// ── Ransomware ──────────────────────────────────────────────────
	`rule Ransomware_WannaCry_Indicators : ransomware {
	meta:
		description = "Detects WannaCry ransomware indicators"
		category = "ransomware"
		severity = "critical"
		confidence = "0.90"
		author = "AxiomNizam"
	strings:
		$s1 = "WannaCrypt"
		$s2 = "WNcry@2ol7"
		$s3 = "msg/m_english.wnry"
		$s4 = "tasksche.exe"
		$s5 = "iuqerfsodp9ifjaposdfjhgosurijfaewrwergwea"
	condition:
		2 of them
}`,

	`rule Ransomware_Generic_Note : ransomware {
	meta:
		description = "Detects generic ransomware note patterns"
		category = "ransomware"
		severity = "critical"
		confidence = "0.80"
		author = "AxiomNizam"
	strings:
		$s1 = "Your files have been encrypted"
		$s2 = "send bitcoin"
		$s3 = "pay ransom"
		$s4 = "decrypt your files"
		$s5 = "bitcoin wallet"
		$s6 = "restore your files"
	condition:
		2 of them
}`,

	`rule Ransomware_LockBit : ransomware {
	meta:
		description = "Detects LockBit ransomware indicators"
		category = "ransomware"
		severity = "critical"
		confidence = "0.88"
		author = "AxiomNizam"
	strings:
		$s1 = "LockBit"
		$s2 = ".lockbit"
		$s3 = "Restore-My-Files.txt"
	condition:
		2 of them
}`,

	// ── Cryptominer ─────────────────────────────────────────────────
	`rule CoinMiner_XMRig : miner {
	meta:
		description = "Detects XMRig cryptocurrency miner"
		category = "cryptominer"
		severity = "high"
		confidence = "0.90"
		author = "AxiomNizam"
	strings:
		$s1 = "xmrig"
		$s2 = "stratum+tcp://"
		$s3 = "stratum+ssl://"
		$s4 = "mining.subscribe"
		$s5 = "randomx"
	condition:
		2 of them
}`,

	`rule CoinMiner_Pool_URLs : miner {
	meta:
		description = "Detects cryptocurrency mining pool URLs"
		category = "cryptominer"
		severity = "high"
		confidence = "0.85"
		author = "AxiomNizam"
	strings:
		$s1 = "pool.minexmr.com"
		$s2 = "monerohash.com"
		$s3 = "moneropool.com"
		$s4 = "pool.supportxmr.com"
		$s5 = "nanopool.org"
		$s6 = "hashvault.pro"
	condition:
		any of them
}`,

	// ── Webshell ────────────────────────────────────────────────────
	`rule Webshell_PHP_Generic : webshell {
	meta:
		description = "Detects common PHP webshell patterns"
		category = "webshell"
		severity = "critical"
		confidence = "0.85"
		author = "AxiomNizam"
	strings:
		$s1 = "eval($_POST["
		$s2 = "eval($_GET["
		$s3 = "eval($_REQUEST["
		$s4 = "eval(base64_decode("
		$s5 = "assert($_POST["
		$s6 = "system($_GET["
		$s7 = "passthru($_"
		$s8 = "shell_exec($_"
	condition:
		any of them
}`,

	`rule Webshell_PHP_Obfuscated : webshell {
	meta:
		description = "Detects obfuscated PHP webshell techniques"
		category = "webshell"
		severity = "critical"
		confidence = "0.82"
		author = "AxiomNizam"
	strings:
		$s1 = "chr(99).chr(104).chr(114)"
		$s2 = "preg_replace"
		$s3 = "create_function"
		$s4 = "call_user_func"
		$s5 = "str_rot13"
		$s6 = "gzinflate(base64_decode("
	condition:
		2 of them
}`,

	`rule Webshell_JSP_Runtime : webshell {
	meta:
		description = "Detects JSP webshell using Runtime.exec"
		category = "webshell"
		severity = "critical"
		confidence = "0.85"
		author = "AxiomNizam"
	strings:
		$s1 = "Runtime.getRuntime().exec"
		$s2 = "ProcessBuilder"
		$s3 = "getParameter"
	condition:
		$s1 and $s3
}`,

	// ── Exploit / Attack Tools ──────────────────────────────────────
	`rule Exploit_Log4Shell : exploit {
	meta:
		description = "Detects Log4Shell JNDI injection payloads"
		category = "exploit"
		severity = "critical"
		confidence = "0.95"
		author = "AxiomNizam"
	strings:
		$s1 = "${jndi:ldap://"
		$s2 = "${jndi:rmi://"
		$s3 = "${jndi:dns://"
		$s4 = "${jndi:iiop://"
	condition:
		any of them
}`,

	`rule Exploit_Reverse_Shell : exploit backdoor {
	meta:
		description = "Detects common reverse shell patterns"
		category = "backdoor"
		severity = "critical"
		confidence = "0.82"
		author = "AxiomNizam"
	strings:
		$s1 = "/dev/tcp/"
		$s2 = "mkfifo /tmp/"
		$s3 = "nc -e /bin/"
		$s4 = "python -c 'import socket"
		$s5 = "perl -e 'use Socket"
	condition:
		any of them
}`,

	`rule Exploit_Cobalt_Strike : exploit {
	meta:
		description = "Detects Cobalt Strike beacon indicators"
		category = "exploit"
		severity = "critical"
		confidence = "0.85"
		author = "AxiomNizam"
	strings:
		$s1 = "beacon.dll"
		$s2 = "beacon.exe"
		$s3 = "ReflectiveLoader"
		$s4 = "IEX (New-Object Net.Webclient).DownloadString"
	condition:
		2 of them
}`,

	// ── Dropper / Loader ────────────────────────────────────────────
	`rule Dropper_PowerShell_Download : dropper {
	meta:
		description = "Detects PowerShell download-and-execute patterns"
		category = "dropper"
		severity = "high"
		confidence = "0.78"
		author = "AxiomNizam"
	strings:
		$s1 = "DownloadString("
		$s2 = "DownloadFile("
		$s3 = "Invoke-WebRequest"
		$s4 = "Start-Process"
		$s5 = "IEX("
		$s6 = "Invoke-Expression"
	condition:
		2 of them
}`,
}

// RegisterBuiltinRules parses and adds all built-in rules to the rule set.
// Returns the number of rules successfully registered.
func RegisterBuiltinRules(rs *RuleSet) int {
	loaded := 0
	for _, src := range builtinRuleSources {
		rule, err := ParseRule(src)
		if err != nil {
			continue
		}
		rs.AddRule(*rule)
		loaded++
	}
	return loaded
}

// BuiltinRuleCount returns the number of built-in YARA rules available.
func BuiltinRuleCount() int {
	return len(builtinRuleSources)
}
