package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	V1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/egeneralov/pv-exporter/internal/dirsize"
	"github.com/egeneralov/pv-exporter/internal/inode"
)

var (
	loglevel      string
	metricsListen string
	kubeconfig    *string
	config        *rest.Config
	clientset     *kubernetes.Clientset

	hostname = os.Getenv("HOSTNAME")
	rootfs   string
	PVS      []V1.PersistentVolume

	err error
)

func main() {
	initApp()

	go func() {
		log.WithFields(log.Fields{
			"metricsListen": metricsListen,
		}).Info("Starting webserver")

		http.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(metricsListen, nil)
	}()

	for {
		time.Sleep(time.Second)
		PVS = gather()

		for _, pv := range PVS {
			log.WithFields(log.Fields{
				"pv": pv.ObjectMeta.Name,
			}).Trace("Already registered")

			pvPath := rootfs + pv.Spec.HostPath.Path

			// available_bytes

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "capacity_bytes",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("capacity_bytes %v", pv.ObjectMeta.Name)
					var stat syscall.Statfs_t
					if err := syscall.Statfs(pvPath, &stat); err != nil {
						log.Error(err)
					}
					return float64(stat.Blocks * uint64(stat.Bsize))
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

			// available_bytes

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "available_bytes",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("available_bytes %v", pv.ObjectMeta.Name)
					var stat syscall.Statfs_t
					if err := syscall.Statfs(pvPath, &stat); err != nil {
						log.Error(err)
					}
					return float64(stat.Bavail * uint64(stat.Bsize))
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

			// used_bytes

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "used_bytes",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("used_bytes %v", pv.ObjectMeta.Name)
					return float64(dirsize.DirSize(pvPath))
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

			// inodes_free

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "inodes_free",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("inodes_free %v", pv.ObjectMeta.Name)
					// itotal, iused, iavail, ipcent, err := inode.GetInodesInfo(path)
					_, _, iavail, _, err := inode.GetInodesInfo(pvPath)
					if err != nil {
						return float64(-1)
					}
					return float64(iavail)
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

			// inodes_used

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "inodes_used",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("inodes_used %v", pv.ObjectMeta.Name)
					// itotal, iused, iavail, ipcent, err := inode.GetInodesInfo(path)
					_, iused, _, _, err := inode.GetInodesInfo(pvPath)
					if err != nil {
						return float64(-1)
					}
					return float64(iused)
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

			// inodes

			if err := prometheus.Register(prometheus.NewGaugeFunc(
				prometheus.GaugeOpts{
					Namespace: "kubelet_volume_stats",
					Name:      "inodes",
					Help:      "",
					ConstLabels: prometheus.Labels{
						"persistentvolume":      pv.ObjectMeta.Name,
						"persistentvolumeclaim": pv.Spec.ClaimRef.Name,
						"namespace":             pv.Spec.ClaimRef.Namespace,
					},
				},
				func() float64 {
					log.Tracef("inodes %v", pv.ObjectMeta.Name)
					// itotal, iused, iavail, ipcent, err := inode.GetInodesInfo(path)
					itotal, _, _, _, err := inode.GetInodesInfo(pvPath)
					if err != nil {
						return float64(-1)
					}
					return float64(itotal)
				},
			)); err != nil {
				log.WithFields(log.Fields{
					"pv": pv.ObjectMeta.Name,
				}).Trace("Already registered")
			}

		}
	}
}

func getSize() string {
	used_bytes := ""

	for _, pv := range PVS {
		used_bytes += fmt.Sprintf(
			"kubelet_volume_stats_used_bytes{persistentvolume=\"%v\",persistentvolumeclaim=\"%v\",namespace=\"%v\"} %d\n",
			pv.ObjectMeta.Name,
			pv.Spec.ClaimRef.Name,
			pv.Spec.ClaimRef.Namespace,
			dirsize.DirSize(rootfs+pv.Spec.HostPath.Path),
		)
	}
	return used_bytes
}

func initApp() {
	inCluster := flag.Bool("in-cluster", false, "is in cluster start")
	flag.StringVar(&hostname, "hostname", hostname, "host to search pv")
	flag.StringVar(&rootfs, "rootfs", "", "prefix for datadir")
	flag.StringVar(&loglevel, "loglevel", "info", "[trace, debug, info]")
	flag.StringVar(&metricsListen, "listen-addr", "0.0.0.0:2112", "bind http server")

	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}

	flag.Parse()

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	switch loglevel {
	case "trace":
		log.SetLevel(log.TraceLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	default:
		log.SetLevel(log.InfoLevel)
	}

	switch *inCluster {
	case true:
		err = inClusterClient()
	case false:
		err = outOfClusterClient()
	}

	if err != nil {
		log.Error(err)
	}
}

func gather() []V1.PersistentVolume {
	log.Debug("gather()")
	var (
		result []V1.PersistentVolume
	)

	pvs, err := clientset.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Error(err)
	}

	for _, pv := range pvs.Items {
		log.Tracef("gather: processing: %v", pv.ObjectMeta.Name)
		// not exist path to dir
		if pv.Spec.HostPath == nil {
			log.Tracef("gather: %v: pv.Spec.HostPath == nil", pv.ObjectMeta.Name)
			continue
		}
		// empty node selector
		if pv.Spec.NodeAffinity == nil {
			log.Tracef("gather: %v: pv.Spec.NodeAffinity == nil", pv.ObjectMeta.Name)
			continue
		}
		if len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms) == 0 {
			log.Tracef("gather: %v: len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms) == 0", pv.ObjectMeta.Name)
			continue
		}
		if len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions) == 0 {
			log.Tracef("gather: %v: len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions) == 0", pv.ObjectMeta.Name)
			continue
		}
		if pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key != "kubernetes.io/hostname" {
			log.Tracef("gather: %v: pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Key != \"kubernetes.io/hostname\"", pv.ObjectMeta.Name)
			continue
		}
		if len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values) == 0 {
			log.Tracef("gather: %v: len(pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values) == 0", pv.ObjectMeta.Name)
			continue
		}
		// not my hostname
		if pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0] != hostname {
			log.Tracef("gather: %v: pv.Spec.NodeAffinity.Required.NodeSelectorTerms[0].MatchExpressions[0].Values[0] != %v", pv.ObjectMeta.Name, hostname)
			continue
		}
		log.Tracef("gather: append: %v", pv.ObjectMeta.Name)
		result = append(result, pv)
	}
	return result
}

func homeDir() string {
	log.Debug("homeDir()")
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

func outOfClusterClient() error {
	log.Debug("outOfClusterClient()")
	config, err = clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}

func inClusterClient() error {
	log.Debug("inClusterClient()")
	config, err = rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}
	return nil
}
