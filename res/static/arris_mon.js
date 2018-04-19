// Arris Monitor Client 2.0
// 2018

/**
 * ArrisMonitor class
 */
function ArrisMonitor(arrisOptions) {
    this.url = arrisOptions.url;    
    this.timeout = arrisOptions.timeout;
    this.downstream = new ArrisStream(arrisOptions.downstream);
    this.upstream = new ArrisStream(arrisOptions.upstream);

    this._paused = false;
    this._started = false;

    if (!arrisOptions.disabled) {
        this._start();
    }

    if (arrisOptions.controls) {
        startButton = document.querySelector(arrisOptions.controls.startButton);
        startButton.textContent = this._paused ? "[Run]" : "[Pause]";
        startButton.addEventListener(
            'click', 
            function(e) {
                e.preventDefault();
                if (!this._paused) {
                    this._stop();
                    e.target.textContent = "[Run]"
                } else {
                    this._start();
                    e.target.textContent = "[Pause]"
                }
            }.bind(this)
        );
    }
}

ArrisMonitor.prototype._start = function () {
    if (!this._started) {
        setInterval(function () {
            if (!this._paused) {
                this.update();
            }
        }.bind(this), this.timeout)
        this._started = true;
    } else if (this._paused) {
        this._paused = false;
    }
}

ArrisMonitor.prototype._stop = function () {
    if (!this._paused) {
        this._paused = true;
    }
}

ArrisMonitor.prototype.update = function () {
    fetch(this.url).then(
        response => {
            if (response.ok) {
                return response.text();
            } else {
                return Promise.reject(response);
            }
        }
    ).then(
        text => this.render(this.parse(text))
    ).catch(
        e => console.log(e)
    )    
}

ArrisMonitor.prototype.render = function (metrics) {    
    if (metrics) {
        var t = new Date().getTime();
        this.downstream.render(t, metrics);
        this.upstream.render(t, metrics);
    }
}

ArrisMonitor.prototype.parse = function (text) {    
    var lines = text.split("\n");    
    var re = /^(\w+)(\{.*\})? (.+)$/u;

    var metrics = {};
    for (var i = 0; i < lines.length; ++i) {
        var line = lines[i].trim();
        if (line.length === 0 || line.startsWith('#')) {
            continue;
        }
        var [_, name, labels, val] = re.exec(line);
        var metric = {
            "val": parseFloat(val), 
            "labels": labels
        };
        if (!metrics[name]) {
            metrics[name] = [];
        }
        metrics[name].push(metric);        
    }    
    return metrics;
}

/**
 * ArrisStream class
 */
function ArrisStream(streamOptions) {
    this.charts = streamOptions.charts;    

    for (var c in this.charts) {
        var chart = this.charts[c];
        this[chart] = {
            chart: new SmoothieChart(streamOptions[chart].chartOptions),
            series: []
        };
        this[chart].chart.streamTo(
            document.querySelector(streamOptions[chart].chartCanvas),
            streamOptions[chart].chartDelay
        );
    }
    for (var i = 0; i < streamOptions.channels.length; i++) {
        for (var c in this.charts) {
            var chart = this.charts[c];
            var ts = new TimeSeries();
            this[chart].series.push(ts);
            this[chart].chart.addTimeSeries(ts, streamOptions.channels[i]);
        }
    }
}

ArrisStream.prototype.render = function (time, metrics) {    
    for (var c in this.charts) {
        var chart = this.charts[c];
        for (var i = 0;  i < this[chart].series.length; i++) {
            try {
                var val = metrics[chart] ? metrics[chart][i].val : this[chart].none;
            }
            catch (e) {
                console.log(e);
            }
            finally {
                if (val) {
                    this[chart].series[i].append(time, val);
                }
            }
        }
    }
}
