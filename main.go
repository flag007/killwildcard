package main

import (
	"github.com/logrusorgru/aurora"
	"fmt"
	"strings"
	"net"
	"os"
	"sync"
	"flag"
	"bufio"
)



type SafeMap struct {
	sync.RWMutex
	Map map[string]int
}

func newSafeMap() *SafeMap {
	sm := new(SafeMap)
	sm.Map = make(map[string]int)
	return sm
}


func (sm *SafeMap) okMap(key string) bool {
	sm.RLock()
	_, ok := sm.Map[key]
	sm.RUnlock()
	return ok
}


func (sm *SafeMap) readMap(key string) int {
	sm.RLock()
	value := sm.Map[key]
	sm.RUnlock()
	return value
}
func (sm *SafeMap) writeMap(key string) {
	sm.Lock()
	value := sm.Map[key]
	sm.Map[key] = value+1
	sm.Unlock()
}

var au aurora.Aurora
var details bool

func init() {
	au = aurora.NewAurora(true)
}

func main() {
	ip_dicc := newSafeMap()

	sc := bufio.NewScanner(os.Stdin)
	var domains []string
	domain_channel1 := make(chan string)
	var domainWG1 sync.WaitGroup

	domain_channel2 := make(chan string)
	var domainWG2 sync.WaitGroup

	output := make(chan string)
	
	var concurrency int
	flag.IntVar(&concurrency, "c", 50, "设置线程")

	flag.BoolVar(&details, "v", false, "输出详情")

	flag.Parse()

	if details {
		str := `

█  █▀ ▄█ █    █      ▄ ▄   ▄█ █     ██▄   ▄█▄    ██   █▄▄▄▄ ██▄   
█▄█   ██ █    █     █   █  ██ █     █  █  █▀ ▀▄  █ █  █  ▄▀ █  █  
█▀▄   ██ █    █    █ ▄   █ ██ █     █   █ █   ▀  █▄▄█ █▀▀▌  █   █ 
█  █  ▐█ ███▄ ███▄ █  █  █ ▐█ ███▄  █  █  █▄  ▄▀ █  █ █  █  █  █  
  █    ▐     ▀    ▀ █ █ █   ▐     ▀ ███▀  ▀███▀     █   █   ███▀  
 ▀                   ▀ ▀                           █   ▀          
                                                  ▀               
						  
		`
		fmt.Println(au.Magenta(str))
		
	}

	for i := 0; i < concurrency; i++ {
		domainWG1.Add(1)

		go func() {
			for domain := range domain_channel1 {
				ip, err := net.ResolveIPAddr("ip", domain)
				if err != nil {
					continue
				}
				ip_dicc.writeMap(ip.String())
			}
			domainWG1.Done()
		}()
	}

	for sc.Scan() {
		domain := strings.ToLower(sc.Text())
		domain_channel1 <- domain
		domains = append(domains, domain)
	}

	if err := sc.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to read input: %s\n", err)
	}

	close(domain_channel1)
	domainWG1.Wait()

	if details {
		
		for key, value := range ip_dicc.Map{
			fmt.Println(au.Yellow(key), au.Yellow(value))
		}

		fmt.Println()
		fmt.Println()
	}

	var inputWG sync.WaitGroup
	inputWG.Add(1)
	go func(){
		for _,domain := range domains {
			domain_channel2 <- domain
		}
		inputWG.Done()
	}()

	go func() {
		inputWG.Wait()
		close(domain_channel2)
	}()

	for i := 0; i < concurrency; i++ {
		domainWG2.Add(1)

		go func() {
			for domain := range domain_channel2 {
				ip, err := net.ResolveIPAddr("ip", domain)
				if err != nil {
					continue
				}

				if ok := ip_dicc.okMap(ip.String()); !ok{
					output <- domain
					continue
				}

				if  ip_dicc.readMap(ip.String()) < 50 {
					output <- domain
				}
			}
			domainWG2.Done()
		}()

	}

	go func() {
		domainWG2.Wait()
		close(output)
	}()

	var outputWG sync.WaitGroup

	outputWG.Add(1)

	go func() {
		for o := range output {
			if details {
				fmt.Println("[!]", o)
			} else {
				fmt.Println(o)
			}
		}
		outputWG.Done()
	}()

	outputWG.Wait()

}
