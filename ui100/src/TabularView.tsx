import {Box, Paper} from "@mui/material";
import useStore from "./model/store.ts";
import {
    MaterialReactTable,
    type MRT_ColumnDef,
    MRT_RowSelectionState,
    useMaterialReactTable
} from "material-react-table";
import {useEffect, useMemo, useState} from "react";
import {Node} from "@xyflow/react";

const TabularView = () => {
    const nodes = useStore((state) => state.nodes);
    const updateNodes = useStore((state) => state.updateNodes);
    const selectedNode = useStore((state) => state.selectedNode);
    const updateSelectedNode = useStore((state) => state.updateSelectedNode);
    const [rowSelection, setRowSelection] = useState<MRT_RowSelectionState>({});

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

    const columns = useMemo<MRT_ColumnDef<Node>[]>(
        () => [
            {
                accessorKey: 'data.label',
                header: 'Label'
            },
            {
                accessorKey: 'type',
                header: 'Type',
            }
        ],
        [],
    );

    const table = useMaterialReactTable({
        columns: columns,
        data: nodes,
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