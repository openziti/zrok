const compareArrays = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e === b[i]);
}

const compareLinks = (a, b) => {
    if(a.length !== b.length) return false;
    return a.every((e, i) => e.source === b[i].source && e.target === b[i].target);
}

export const mergeGraph = (oldGraph, newOverview) => {
    let newGraph = {
        nodes: [],
        links: []
    }
    newOverview.forEach(env => {
        newGraph.nodes.push({
            id: env.environment.zId,
            label: env.environment.description,
            type: "environment"
        });
        if(env.services) {
            env.services.forEach(svc => {
                let svcLabel = svc.token;
                if(svc.backendProxyEndpoint !== "") {
                    svcLabel = svc.backendProxyEndpoint;
                }
                newGraph.nodes.push({
                    id: svc.token,
                    label: svcLabel,
                    type: "service",
                    val: 10
                });
                newGraph.links.push({
                    target: env.environment.zId,
                    source: svc.token,
                    color: "#777"
                });
            });
        }
    });
    // we want to preserve nodes that exist in the new graph, and remove those that don't.
    let outputNodes = oldGraph.nodes.filter(oldNode => newGraph.nodes.find(newNode => newNode.id === oldNode.id));
    let outputLinks = oldGraph.nodes.filter(oldLink => newGraph.links.find(newLink => newLink.target === oldLink.target && newLink.source === oldLink.source));
    // and then do the opposite; add any nodes that are in newGraph that are missing from oldGraph.
    outputNodes.push(...newGraph.nodes.filter(newNode => !outputNodes.find(oldNode => oldNode.id === newNode.id)));
    outputLinks.push(...newGraph.links.filter(newLink => !outputLinks.find(oldLink => oldLink.target === newLink.target && oldLink.source === newLink.source)));
    outputNodes = outputNodes.sort();
    outputLinks = outputLinks.sort();
    return {
        nodes: outputNodes,
        links: outputLinks
    };
};