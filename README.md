# Arris Monitor

Streaming ARRIS CM820B Cable Modem data into live charts.

## Usage

```
> mkdir -p $(GOPATH)/src/github.com/realitycheck/
> cd $(GOPATH)/src/github.com/realitycheck/
> git clone https://github.com/realitycheck/arris_mon
> cd arris_mon
> make install
> arris_mon
# Open http://127.0.0.1:8081 in your browser
```

![Arris Monitor](arris_mon.png)

## Prometheus

Prometheus metrics are exported by `http://127.0.0.1:8081/metrics`

## System

* ARRIS EuroDOCSIS 3.0 Touchstone WideBand Cable Modem
* MODEL: CM820B
* Firmware Name:	TS0705125_062314_WBM760_CM820