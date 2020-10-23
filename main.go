package main

import (
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type msg struct {
	Type    string
	Content map[string]string
}

// KubeConfig define config file path
type kubeConfig struct {
	Path string
}

func main() {
	m := make(map[string]string)

	mux := &sync.RWMutex{}
	wg := &sync.WaitGroup{}
	ch := make(chan msg)
	defer close(ch)
	go liveLoop(ch)
	go writeLoop(m, mux, ch)
	go readLoop(m, mux, wg)
	// stop program from exiting, must be killed
	printer := make(chan struct{})
	<-printer
}

func writeLoop(m map[string]string, mux *sync.RWMutex, ch <-chan msg) {
	//fmt.Printf("event: %s\n", a)
	//fmt.Printf("map: %s\n", m)
	for {
		a := <-ch
		switch a.Type {
		case "ADDED":
			//fmt.Println("ADDED")
			for k, v := range a.Content {
				fmt.Printf("%s was added to the detection of the lists\n", k)
				mux.Lock()
				m[k] = v
				mux.Unlock()
			}
		case "DELETED":
			for k, _ := range a.Content {
				fmt.Printf("%s was deleted from the detection of the lists\n", k)
				mux.Lock()
				delete(m, k)
				mux.Unlock()
			}
			fmt.Println(m)
		default:
			fmt.Println("The Service is being monitores....")
		}
	}
}

func readLoop(m map[string]string, mux *sync.RWMutex, wg *sync.WaitGroup) {
	for {
		mux.RLock()
		for k, v := range m {
			url := v
			// fmt.Println(url)
			wg.Add(1)
			go func(k, url string) {
				defer wg.Done()
				_, err := http.Get(m[k])
				if err != nil {
					fmt.Println(k, err)
				} else {
					fmt.Println(k + " is ok...")
				}

			}(k, url)
		}
		wg.Wait()
		mux.RUnlock()
		time.Sleep(5 * time.Second)
	}
}

func liveLoop(ch chan msg) {
	var k kubeConfig
	k.Path = "/Users/weeknd/Desktop/k8s/config"
	clientset := k.NewClient()
	var dst = msg{}
	fmt.Println("Living loop Now")

	watcher, err := clientset.CoreV1().Services("rw-product").Watch(metav1.ListOptions{LabelSelector: "project=rw"})
	if err != nil {
		fmt.Printf("err: %s\n", err)
	}
	for {
		event := <-watcher.ResultChan()
		s := event.Object.(*v1.Service)
		//fmt.Println(event)
		switch event.Type {
		case "ADDED":
			dst.Type = "ADDED"
			data := make(map[string]string)
			data[s.Name] = "https://" + s.Name + ".wukongbox.co"
			dst.Content = data
			ch <- dst
		case "DELETED":
			dst.Type = "DELETED"
			data := make(map[string]string)
			data[s.Name] = "https://" + s.Name + ".wukongbox.co"
			dst.Content = data
			ch <- dst
		}
	}
}

func (c *kubeConfig) NewClient() *kubernetes.Clientset {

	config, err := clientcmd.BuildConfigFromFlags("", c.Path)
	if err != nil {
		panic(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err)
	}
	return clientset
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func int32Ptr(i int32) *int32 {
	return &i
}
