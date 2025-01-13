import {Box, Paper} from "@mui/material";
import useStore from "./model/store.ts";
import {
    MaterialReactTable,
    type MRT_ColumnDef,
    MRT_RowSelectionState,
    useMaterialReactTable
} from "material-react-table";
import {useEffect, useMemo, useRef, useState} from "react";
import {Node} from "@xyflow/react";
import {bytesToSize} from "./model/util.ts";

const TabularView = () => {
    const nodes = useStore((state) => state.nodes);
    const nodesRef = useRef<Node[]>();
    nodesRef.current = nodes;
    const updateNodes = useStore((state) => state.updateNodes);
    const selectedNode = useStore((state) => state.selectedNode);
    const updateSelectedNode = useStore((state) => state.updateSelectedNode);
    const [rowSelection, setRowSelection] = useState<MRT_RowSelectionState>({});
    const sparkdata = useStore((state) => state.sparkdata);
    const [combined, setCombined] = useState<Node[]>([]);

    useEffect(() => {
        let outNodes = new Array<Node>();
        nodesRef.current.forEach(node => {
            let outNode = {
                ...node
            };
            outNode.data.activity = sparkdata.get(node.id);
            outNodes.push(outNode);
        });
        setCombined(outNodes);
    }, [nodes, sparkdata]);

    useEffect(() => {
        if(selectedNode) {
            let selection = {};
            selection[selectedNode.id] = true;
            setRowSelection(selection);
        }
    }, []);

    useEffect(() => {
        let sn = nodes.find(node => Object.keys(rowSelection).includes(node.id));
        updateSelectedNode(sn);
        updateNodes(nodes.map(node => (sn && node.id === sn.id) ? { ...node, selected: true } : { ...node, selected: false }));
    }, [rowSelection]);

    const sparkdataTip = (row) => {
        if(row.data.activity) {
            let tip = row.data.activity[row.data.activity.length - 1];
            if(tip > 0) {
                return bytesToSize(tip);
            }
        } else {
            console.log("no sparkdata", row);
        }
        return "";
    };

    const columns = useMemo<MRT_ColumnDef<Node>[]>(
        () => [
            {
                accessorKey: 'data.label',
                header: 'Label'
            },
            {
                accessorKey: 'type',
                header: 'Type',
            },
            {
                accessorFn: sparkdataTip,
                header: 'Activity',
            }
        ],
        [],
    );

    const table = useMaterialReactTable({
        columns: columns,
        data: combined,
        enableRowSelection: false,
        enableMultiRowSelection: false,
        getRowId: r => r.id,
        onRowSelectionChange: setRowSelection,
        state: { rowSelection },
        muiTableBodyRowProps: ({ row }) => ({
            onClick: () => {
                if(rowSelection[row.id]) {
                    setRowSelection({});
                } else {
                    setRowSelection({[row.id]: true});
                }

            },
            selected: rowSelection[row.id],
            sx: {
                cursor: 'pointer',
            },
        }),
        muiToolbarAlertBannerProps: {
            sx: {
                color: "#241775",
                backgroundColor: "#f5fde7",
            }
        },
        positionToolbarAlertBanner: "bottom",
    });

    return (
        <Box sx={{ width: "100%", mt: 2 }} height={{ xs: 400, sm: 600, md: 800 }}>
            <Paper>
                <MaterialReactTable table={table} />
            </Paper>
        </Box>
    );
};

export default TabularView;