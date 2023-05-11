export const buildMetrics = (m) => {
    console.log("build", m);
    let metrics = {
        data: m.samples,
        rx: 0,
        tx: 0
    }
    if(m.samples) {
        m.samples.forEach(s => {
            metrics.rx += s.rx ? s.rx : 0;
            metrics.tx += s.tx ? s.tx : 0;
        });
    }
    return metrics;
}

export const bytesToSize = (sz) => {
    let absSz = sz;
    if(absSz < 0) {
        absSz *= -1;
    }
    const unit = 1000
    if(absSz < unit) {
        return '' + absSz + ' B';
    }
    let div = unit
    let exp = 0
    for(let n = absSz / unit; n >= unit; n /= unit) {
        div *= unit;
        exp++;
    }

    return '' + (sz / div).toFixed(1) + ' ' + "kMGTPE"[exp] + 'B';
}