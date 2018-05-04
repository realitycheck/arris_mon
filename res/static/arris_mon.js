// Arris Monitor Client 2.0
// 2018


var arris_mon = (function (){
    return {
        _m: null,
        init: function (opts) {
            if (typeof(opts) === "string") {
                fetch(opts).then(
                    response => response.json()
                ).then(
                    options => {
                        this._m = new ArrisMonitor(options);
                    }
                ).catch(
                    e => console.error(e)
                );
            }            
        }
    };
}());

/**
 * ArrisMonitor class
 */
function ArrisMonitor(arrisOptions) {    
    this.url = arrisOptions.url;    
    this.timeout = arrisOptions.timeout;
    this.graphs = {};
    
    this._paused = false;
    this._started = false;    

    if (!arrisOptions.disabled) {
        this._start();
    }

    if (arrisOptions.controls) {
        var startButton = document.querySelector(arrisOptions.controls.startButton);
        startButton.textContent = this._paused ? "[Run]" : "[Pause]";
        startButton.addEventListener(
            "click", function(e) {
                e.preventDefault();
                if (!this._paused) {
                    this._stop();
                    e.target.textContent = "[Run]";
                } else {
                    this._start();
                    e.target.textContent = "[Pause]";
                }
            }.bind(this)
        );
    }

    if (arrisOptions.graphs) {
        for (var m in arrisOptions.graphs) {
            this.graphs[m] = new ArrisGraph(m, arrisOptions.graphs[m]);            
        }
    }
}

ArrisMonitor.prototype._start = function () {
    if (!this._started) {
        setInterval(function () {
            if (!this._paused) {
                this.update();
            }
        }.bind(this), this.timeout);
        this._started = true;
    } else if (this._paused) {
        this._paused = false;
    }
};

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
        e => console.warn(e)
    );
};

ArrisMonitor.prototype.render = function (metrics) {    
    if (metrics) {
        var t = new Date().getTime();
        
        for (var g in this.graphs) {
            if (!metrics[g]) {
                // console.debug("no graph metric:", g);
                continue;
            }
            
            for (var i in metrics[g]) {
                this.graphs[g].renderMetric(t, metrics[g][i]);
            }

            this.graphs[g].renderLegend();
        }
    }
};

ArrisMonitor.prototype.parse = function (text) {    
    var lines = text.split("\n");    
    var re = /^(\w+)(\{.*\})? (.+)$/u;

    var metrics = {};
    for (var i = 0; i < lines.length; ++i) {
        var line = lines[i].trim();
        if (line.length === 0 || line.startsWith("#")) {
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
};


function ArrisGraph(name, graphOptions) {
    this.name = name;
    this.options = graphOptions;

    this.chart = new SmoothieChart(this.options.chartOptions);
    this.series = {};
    this.chart.streamTo(
        document.querySelector(this.options.chartCanvas),
        this.options.chartDelay
    );

    this.legend = null;
    if (this.options.legendOptions) {
        this.legend = document.querySelector(this.options.legendOptions.selector);
    }
} 

ArrisGraph.prototype.renderMetric = function (time, metric) {
    var series = this.series[metric.labels];
    if (!series) {
        var timeOptions = patternOptions(this.options.timeOptions, metric.labels);
        var seriesOptions = patternOptions(this.options.seriesOptions, metric.labels);

        var ts = new TimeSeries(timeOptions);        
        this.chart.addTimeSeries(ts, seriesOptions);
        series = this.series[metric.labels] = {
            ts: ts,
            options: this.chart.seriesSet[this.chart.seriesSet.length-1].options,
        };
    }

    // console.debug("render metric:", metric.val, metric.labels);
    if (metric.val) {
        series.ts.append(time, metric.val);
    }    
};

ArrisGraph.prototype.renderLegend = function () {
    if (!this.legend) {
        return;
    }

    var innerHTML = [
        "<table>" + "<caption>" + this.name + "</caption>"
    ];
    for (var s in this.series) {
        var series = this.series[s];
        var box = [
            "<div style='" + "width: 12px; height: 4px;",
            "background: " + series.options.strokeStyle + ";'></div>",
        ].join(" ");
        
        var i = series.ts.data.length - 1;
        innerHTML.push(
            "<tr>",            
            "<td>", s, "</td>",
            "<td>", box, "</td>",
            "<td>", series.ts.data[i][1], "</td>",
            "</tr>"
        );
    }
    innerHTML.push("</table>");

    this.legend.innerHTML = innerHTML.join("");
};

function patternOptions(options, match) {
    if (options) {        
        var re = /^\/\/?(.+)\/\/?(.+)?$/u;
        for (var pattern in options) {
            var [_, regex, flags] = re.exec(pattern);
            var pattern_re = new RegExp(regex, flags);
            if (pattern_re.test(match)) {
                return options[pattern];
            }
        }
    }
    return {};
}