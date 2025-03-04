export const objectToRows = (obj) => {
    let rows = [];
    let count = 0;
    for(const key in obj) {
        rows.push({
            id: count++,
            property: key,
            value: obj[key]
        });
    }
    rows.sort((a, b) => a.property.localeCompare(b.property));
    return rows;
};

export const camelToWords = (s) => s.replace(/([A-Z])/g, ' $1').replace(/^./, function(str){ return str.toUpperCase(); });

export const bytesToSize = (bytes: number): string => {
    let i = -1;
    const byteUnits = [' kB', ' MB', ' GB', ' TB', 'PB', 'EB', 'ZB', 'YB'];
    do {
        bytes /= 1024;
        i++;
    } while (bytes > 1024);
    return Math.max(bytes, 0.1).toFixed(1) + byteUnits[i];
}

export const buildMetrics = (m) => {
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