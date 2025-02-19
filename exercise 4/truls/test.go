package main

import (
	"flag"
	"fmt"
	"net"
	"os/exec"
	"os"
	"strconv"
	"strings"
	"time"
	"runtime"
)

const (
	udpPort = 3000 // trenger egentlig bare å være port mellom 1024 - 49151, altså frie porter
	heartbeatInterval = 1 * time.Second // blir altså i mellom hver print out av tall
	timeoutDuration   = 3 * time.Second // når vi starter primary backup tar vi litt ventetid
)

func main() {
	rolePtr := flag.String("role", "primary", "Role: primary or backup")
	startCountPtr := flag.Int("start", 1, "Starting count")
	flag.Parse()

	if *rolePtr == "primary" {
		runPrimary(*startCountPtr)
	} else if *rolePtr == "backup" {
		runBackup(*startCountPtr)
	} else {
		fmt.Println("Unknown role. Use -role=primary or -role=backup")
	}
}



// runPrimary gjlr:
//  den "Spawner" backup prosessen, med funksjonen: spawnBackup
// Opens a UDP connection to send heartbeat messages.
// printer tall inkrementerende og sender også denne tellingen som heartbeat til porten som bakcup lytter på
func runPrimary(startCount int) {
	fmt.Println("Running as primary")
	spawnBackup(startCount - 1) // Start backup før UDP-tilkobling

	// Ny pause for å sikre at backupen starter riktig
	time.Sleep(2 * time.Second) 

	// starerr UDP-forbindelse
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", udpPort)) //127.0.0.1 loopback-adressen, som betyr at kommunikasjonen skjer lokalt på samme maskin, alternativ til IP
	if err != nil {
		panic(err)
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		panic(err)
	}

	defer conn.Close()

	count := startCount
	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	for {
		<-ticker.C
		message := fmt.Sprintf("%d", count)
		_, err := conn.Write([]byte(message))
		if err != nil {
			fmt.Println("Error sending heartbeat:", err)
		}

		fmt.Println(count)
		count++
	}
}

// runBackup hører etter "heartbeats" på port 30000 
// Om den stopper å få heartbeats etter litt tid --> blir den nye primary og tar over 
func runBackup(startCount int) {
	fmt.Println("Running as backup, waiting for heartbeats...")
	// hører via UDP, port 3000
	addr, err := net.ResolveUDPAddr("udp", fmt.Sprintf(":%d", udpPort))
	if err != nil {
		panic(err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	lastCount := startCount
	buf := make([]byte, 1024)
	for {
		// Set a read deadline so that we don't block forever.
		conn.SetReadDeadline(time.Now().Add(timeoutDuration))
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			// får vi timeout noen gang, antar vi alltid at primary er død 
			fmt.Println("No heartbeat received; primary appears dead. Promoting to primary...")
			conn.Close() // denne er for å close connection fordi port 3000 allerede er opptatt, slik at "drept" prosess ikke okkuperer porten 
			// blr til Primary (fortsetter telling med lastCount + 1).
			runPrimary(lastCount + 1)
			return
		}

		// tolke og konvertere meldingen som backupen mottar fra primary
		msg := strings.TrimSpace(string(buf[:n]))
		newCount, err := strconv.Atoi(msg) // konverter heartbeat-meldingen (som er et tall i strengformat) til en int
		if err != nil {
			fmt.Println("Error parsing heartbeat message:", err)
			continue
		}
		lastCount = newCount
		
	}
}


// spawnBackup funksjonen åpner et nytt terminal vindu for da enten macOS eller linux (så lengr GNOME terminal er brukt),
// i denne terminalen kjøres en backup prosess som leser av senste telling på porten 3000 og lytter for nye slike heartbeats hetl
// til den oppdager et 3 sekunders mellomrom (skjer når jeg dreper pirmær prosessen) og starter da backupen. 
func spawnBackup(lastCount int) {
    var cmd *exec.Cmd

    // Hent absolutt sti til programmet
    exePath, err := os.Getwd()
    if err != nil {
        fmt.Println("Error getting working directory:", err)
        return
    }

    if runtime.GOOS == "darwin" { // macOS
        cmd = exec.Command("osascript", "-e", fmt.Sprintf(
            `tell app "Terminal" to do script "cd %s && ./test -role=backup -start=%d"`, exePath, lastCount))
    } else if runtime.GOOS == "linux" { // Linux
		exePath, err = os.Getwd()
//        cmd = exec.Command("kgx", "--", "bash", "-c", "cd ", exePath, "&& ./test -role=backup")
		cmd = exec.Command("kgx", "-c", fmt.Sprintf("kgx ./test -role=backup"), exePath)
    } else {
        fmt.Println("Unsupported OS")
        return
    }

    err = cmd.Start()
    if err != nil {
        fmt.Println("Error spawning backup process:", err)
    }
}




