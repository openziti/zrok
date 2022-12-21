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

const sortLinks = (links) => {
    return links.sort((a, b) => {
        if(a.target.id+a.source.id > b.target.id+b.source.id) {
            return 1;
        }
        if(a.target.id+a.source.id < b.target.id+b.source.id) {
            return -1;
        }
        return 0;
    })
}

const graphsEqual = (a, b) => {
    let nodes = nodesEqual(a.nodes, b.nodes);
    let links = linksEqual(a.links, b.links);
    if(!links) {
        console.log("a", a.links.map(l => l.target + ' ' + l.source));
        console.log("b", b.links.map(l => l.target + ' ' + l.source))
    }
    return nodes && links;
}

const nodesEqual = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.id === b[i].id);
}

const linksEqual = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.target === b[i].target.id && e.source === b[i].source.id);
}

export const mergeGraph = (oldGraph, newOverview) => {
    let newGraph = {
        nodes: [],
        links: []
    }
    newOverview.forEach(env => {
        let envNode = {
            id: env.environment.zId,
            label: env.environment.description,
            type: "environment"
        };
        newGraph.nodes.push(envNode);
        if(env.services) {
            env.services.forEach(svc => {
                let svcLabel = svc.token;
                if(svc.backendProxyEndpoint !== "") {
                    svcLabel = svc.backendProxyEndpoint;
                }
                let svcNode = {
                    id: svc.token,
                    label: svcLabel,
                    type: "service",
                    val: 10
                };
                newGraph.nodes.push(svcNode);
                newGraph.links.push({
                    target: envNode.id,
                    source: svcNode.id,
                    color: "#777"
                });
            });
        }
    });
    newGraph.nodes = sortNodes(newGraph.nodes);
    newGraph.links = sortLinks(newGraph.links);

    if(graphsEqual(newGraph, oldGraph)) {
        return oldGraph;
    }

    // we want to preserve nodes that exist in the new graph, and remove those that don't.
    let outputNodes = oldGraph.nodes.filter(oldNode => newGraph.nodes.find(newNode => newNode.id === oldNode.id));
    let outputLinks = oldGraph.nodes.filter(oldLink => newGraph.links.find(newLink => newLink.target === oldLink.target && newLink.source === oldLink.source));
    // and then do the opposite; add any nodes that are in newGraph that are missing from oldGraph.
    outputNodes.push(...newGraph.nodes.filter(newNode => !outputNodes.find(oldNode => oldNode.id === newNode.id)));
    outputLinks.push(...newGraph.links.filter(newLink => !outputLinks.find(oldLink => oldLink.target === newLink.target && oldLink.source === newLink.source)));
    return {
        nodes: outputNodes,
        links: outputLinks
    };
};