const buildOverview = (status) => {
    let overview = new Map();
    status.accesses.map(acc => {
        overview.set(acc.frontendToken, {
           type: "access",
           v: acc
        });
    });
    status.shares.map(shr => {
        overview.set(shr.token, {
            type: "share",
            v: shr
        })
    });
    return overview;
}

export default buildOverview;