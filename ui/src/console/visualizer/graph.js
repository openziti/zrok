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
    oldGraph.nodes = oldGraph.nodes.filter(oldNode => newGraph.nodes.find(newNode => newNode.id === oldNode.id));
    // and then do the opposite; add any nodes that are in newGraph that are missing from oldGraph.
    oldGraph.nodes.push(...newGraph.nodes.filter(newNode => !oldGraph.nodes.find(oldNode => oldNode.id === newNode.id)));
    return {
        nodes: oldGraph.nodes,
        links: newGraph.links,
    };
};