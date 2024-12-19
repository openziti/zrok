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
    return rows;
};

export const camelToWords = (s) => s.replace(/([A-Z])/g, ' $1').replace(/^./, function(str){ return str.toUpperCase(); });

export const rowToValue = (row) => {
    if(row.property.endsWith("At")) {
        return new Date(row.value).toLocaleString();
    }
    return row.value.toString();
};