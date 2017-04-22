// Arris Monitor 1.0
// 2017

function ArrisMonitor(arrisOptions) {
    this.url = arrisOptions.url;
    this.dataType = "html";
    this.timeout = arrisOptions.timeout;
    this.downstream = new ArrisStream(arrisOptions.downstream);
    this.upstream = new ArrisStream(arrisOptions.upstream);

    this._paused = false;
    this._started = false;

    if (!arrisOptions.disabled) {
        this._start();
    }

    if (arrisOptions.controls) {
        $(arrisOptions.controls.statusButton).attr("href", this.url);
        $(arrisOptions.controls.startButton).text(
            this._paused ? "[Run]" : "[Pause]"
        ).click(function (e) {
            e.preventDefault();
            if (!this._paused) {
                this._stop();
                $(e.target).text("[Run]")
            } else {
                this._start();
                $(e.target).text("[Pause]");
            }
        }.bind(this));
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
    $.ajax({
        url: this.url,
        timeout: this.timeout,
        dataType: this.dataType,
        success: function(response, status, xhr) {
            this.render(response);
        }.bind(this),
        error: function(xhr, status, e) {
            console.log(status, e);
        }.bind(this)
    });
}

ArrisMonitor.prototype.render = function (response) {
    var t = new Date().getTime();
    if (response) {
        this.downstream.render(t, this.downstream.parse(response));
        this.upstream.render(t, this.upstream.parse(response));
    }
}

/**
 */
function ArrisStream(streamOptions) {
    this.charts = streamOptions.charts;
    this.parser = streamOptions.parser;

    for (var c in this.charts) {
        var chart = this.charts[c];
        this[chart] = {
            chart: new SmoothieChart(streamOptions[chart].chartOptions),
            series: []
        };
        this[chart].chart.streamTo($(streamOptions[chart].chartCanvas).get(0),
                                   streamOptions[chart].chartDelay);
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

ArrisStream.prototype.render = function (time, data) {
    for (var c in this.charts) {
        var chart = this.charts[c];
        for (var i = 0;  i < this[chart].series.length; i++) {
            try {
                var val = data ? data[chart][i] : this[chart].none;
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

ArrisStream.prototype.parse = function (status) {
    var streamElement = $(this.parser.streamElement, status).get(0);
    if (!streamElement) {
        return [];
    }
    var data = $(this.parser.dataElement, streamElement).map(function (i, v) {
        return this.parser.parse($(this.parser.parseElement, v));
    }.bind(this)).get();

    result = {};
    for (var i = 1; i < data.length; i++) {
        for (var chart in data[i]) {
            if (!(chart in result)) {
                result[chart] = [];
            }
            result[chart].push(data[i][chart]);
        }
    }
    return result;
}