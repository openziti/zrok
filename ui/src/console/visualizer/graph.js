const sortNodes = (nodes) => {
    return nodes.sort((a, b) => {
        if(a.id > b.id) {
            return 1;
        }
        if(a.id < b.id) {
            return -1;
        }
        return 0;
    })
}

const nodesEqual = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.id === b[i].id && e.limited === b[i].limited && e.label === b[i].label);
}

export const mergeGraph = (oldGraph, user, accountLimited, newOverview) => {
    let newGraph = {
        nodes: [],
        links: []
    }

    let accountNode = {
        id: user.token,
        label: user.email,
        type: "account",
        limited: !!accountLimited,
        val: 50
    }
    newGraph.nodes.push(accountNode);

    if(newOverview) {
        let allShares = {};
        let allFrontends = [];
        newOverview.forEach(env => {
            let envNode = {
                id: 'env:' + env.environment.zId,
                envZId: env.environment.zId,
                label: env.environment.description,
                type: "environment",
                limited: !!env.environment.limited || accountNode.limited,
                val: 50
            };
            newGraph.nodes.push(envNode);
            newGraph.links.push({
                target: accountNode.id,
                source: envNode.id,
                color: "#04adef"
            });
            if(env.shares) {
                env.shares.forEach(shr => {
                    let shrLabel = shr.token;
                    if(shr.backendProxyEndpoint !== "") {
                        shrLabel = shr.backendProxyEndpoint;
                    }
                    let shrNode = {
                        id: 'shr:' + shr.token,
                        shrToken: shr.token,
                        envZId: env.environment.zId,
                        label: shrLabel,
                        type: "share",
                        limited: !!shr.limited || envNode.limited,
                        val: 50
                    };
                    allShares[shr.token] = shrNode;
                    newGraph.nodes.push(shrNode);
                    newGraph.links.push({
                        target: envNode.id,
                        source: shrNode.id,
                        color: "#04adef"
                    });
                });
            }
            if(env.frontends) {
                env.frontends.forEach(fe => {
                   let feNode = {
                       id: 'ac:' + fe.id,
                       feId: fe.id,
                       target: fe.shrToken,
                       label: fe.token,
                       type: "frontend",
                       val: 50
                   }
                   allFrontends.push(feNode);
                   newGraph.nodes.push(feNode);
                   newGraph.links.push({
                       target: envNode.id,
                       source: feNode.id,
                       color: "#04adef"
                   });
                });
            }
        });
        allFrontends.forEach(fe => {
            let target = allShares[fe.target];
            if(target) {
                newGraph.links.push({
                    target: target.id,
                    source: fe.id,
                    color: "#9BF316",
                    type: "data",
                });
            }
        });
    }
    newGraph.nodes = sortNodes(newGraph.nodes);

    if(nodesEqual(oldGraph.nodes, newGraph.nodes)) {
        // if the list of nodes is equal, the graph hasn't changed; we can just return the oldGraph and save the
        // physics headaches in the visualizer.
        return oldGraph;
    }

    // we're going to need to recompute a new graph... but we want to maintain the instances that already exist...

    // we want to preserve nodes that exist in the new graph, and remove those that don't.
    let outputNodes = oldGraph.nodes.filter(oldNode => newGraph.nodes.find(newNode => newNode.id === oldNode.id && newNode.limited === oldNode.limited && newNode.label === oldNode.label));
    let outputLinks = oldGraph.nodes.filter(oldLink => newGraph.links.find(newLink => newLink.target === oldLink.target && newLink.source === oldLink.source));

    // and then do the opposite; add any nodes that are in newGraph that are missing from oldGraph.
    outputNodes.push(...newGraph.nodes.filter(newNode => !outputNodes.find(oldNode => oldNode.id === newNode.id && oldNode.limited === newNode.limited && oldNode.label === newNode.label)));
    outputLinks.push(...newGraph.links.filter(newLink => !outputLinks.find(oldLink => oldLink.target === newLink.target && oldLink.source === newLink.source)));

    return {
        // we need a new outer object, to trigger react to refresh the view.
        nodes: sortNodes(outputNodes),
        links: outputLinks
    };
};