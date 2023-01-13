export const isSelectionGone = (networkGraph, selection) => {
    // the selection is gone if the selection is not found in the network graph
    return !networkGraph.nodes.find(node => selection.id === node.id);
}

export const markSelected = (networkGraph, selection) => {
    networkGraph.nodes.forEach(node => { node.selected = node.id === selection.id; });
}