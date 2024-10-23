const buildOverview = (status) => {
    let overview = [];
    if(status) {
        if(status.accesses) {
            status.accesses.forEach(acc => {
                overview.push({
                    type: "access",
                    v: structuredClone(acc)
                });
            });
        }
        if(status.shares) {
            status.shares.forEach(shr => {
                overview.push({
                    type: "share",
                    v: structuredClone(shr)
                });
            });
        }
    }
    return overview;
}

export default buildOverview;