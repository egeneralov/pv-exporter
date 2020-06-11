patch:
	sed -i 's/MustRegister(NewGoCollector())//g' vendor/github.com/prometheus/client_golang/prometheus/registry.go
	sed -i 's/MustRegister(NewProcessCollector(ProcessCollectorOpts{}))//g' vendor/github.com/prometheus/client_golang/prometheus/registry.go

