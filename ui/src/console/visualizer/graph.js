export const mergeGraph = (oldGraph, newOverview) => {
    let graph = {
        nodes: [],
        links: []
    }
    newOverview.forEach(env => {
        graph.nodes.push({
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
                graph.nodes.push({
                    id: svc.token,
                    label: svcLabel,
                    type: "service",
                    val: 10
                });
                graph.links.push({
                    target: env.environment.zId,
                    source: svc.token,
                    color: "#777"
                });
            });
        }
    });
    graph.nodes.forEach(newNode => {
        let found = oldGraph.nodes.find(oldNode => oldNode.id === newNode.id);
        if(found) {
            newNode.vx = found.vx;
            newNode.vy = found.vy;
            newNode.x = found.x;
            newNode.y = found.y;
        }
    })
    return graph;
};