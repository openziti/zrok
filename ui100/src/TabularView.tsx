import {Box, Paper} from "@mui/material";
import useApiConsoleStore from "./model/store.ts";
import {
    MaterialReactTable,
    type MRT_ColumnDef,
    MRT_PaginationState,
    MRT_RowSelectionState,
    MRT_SortingState,
    useMaterialReactTable
} from "material-react-table";
import {useEffect, useMemo, useRef, useState} from "react";
import {Node} from "@xyflow/react";
import {bytesToSize} from "./model/util.ts";

const TabularView = () => {
    const nodes = useApiConsoleStore((state) => state.nodes);
    const nodesRef = useRef<Node[]>();
    nodesRef.current = nodes;
    const updateNodes = useApiConsoleStore((state) => state.updateNodes);
    const selectedNode = useApiConsoleStore((state) => state.selectedNode);
    const updateSelectedNode = useApiConsoleStore((state) => state.updateSelectedNode);
    const sparkdata = useApiConsoleStore((state) => state.sparkdata);
    const storedPagination = useApiConsoleStore((state) => state.pagination);
    const updatePagination = useApiConsoleStore((state) => state.updatePagination);
    const storedSorting = useApiConsoleStore((state) => state.sorting);
    const updateSorting = useApiConsoleStore((state) => state.updateSorting);
    const [pagination, setPagination] = useState<MRT_PaginationState>({} as MRT_PaginationState);
    const [rowSelection, setRowSelection] = useState<MRT_RowSelectionState>({});
    const [sorting, setSorting] = useState<MRT_SortingState>([{id: "data.label", desc: false}] as MRT_SortingState);
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
        setPagination(storedPagination);
        setSorting(storedSorting);
    }, []);

    useEffect(() => {
        updatePagination(pagination);
    }, [pagination]);

    useEffect(() => {
        updateSorting(sorting);
    }, [sorting]);

    useEffect(() => {
        let sn = nodes.find(node => Object.keys(rowSelection).includes(node.id));
        updateSelectedNode(sn);
        updateNodes(nodes.map(node => (sn && node.id === sn.id) ? { ...node, selected: true } : { ...node, selected: false }));
    }, [rowSelection]);

    const sparkdataTip = (row) => {
        if(row.data && row.data.activity) {
            // - 2; - 1 is sometimes undefined?
            return row.data.activity[row.data.activity.length - 2];
        }
        return 0;
    }

    const sparkdataTipFmt = (row) => {
        let tip = sparkdataTip(row);
        if(tip > 0) {
            return bytesToSize(tip);
        }
        return "";
    };

    const sparkdataAverage = (row) => {
        if(row.data && row.data.activity) {
            let average = row.data.activity.reduce((acc, curr) => { return acc + curr }, 0);
            average /= row.data.activity.length;
            return average;
        }
        return 0;
    }

    const sparkdataAverageFmt = (row) => {
        let average = sparkdataAverage(row);
        if(average > 0) {
            return bytesToSize(average);
        }
        return "";
    }

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
                accessorFn: sparkdataTipFmt,
                header: 'Activity',
                sortingFn: (rowA, rowB) => {
                    let tipA = sparkdataTip(rowA.original);
                    let tipB = sparkdataTip(rowB.original);
                    return tipA > tipB ? 1 : tipA < tipB ? -1 : 0;
                },
                sortDescFirst: true
            },
            {
                accessorFn: sparkdataAverageFmt,
                header: 'Activity 5m',
                sortingFn: (rowA, rowB) => {
                    let avgA = sparkdataAverage(rowA.original);
                    let avgB = sparkdataAverage(rowB.original);
                    return avgA > avgB ? 1 : avgA < avgB ? -1 : 0;
                },
                sortDescFirst: true
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
        onPaginationChange: setPagination,
        onRowSelectionChange: setRowSelection,
        onSortingChange: setSorting,
        state: { pagination, rowSelection, sorting },
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
        mrtTheme: (theme) => ({
            matchHighlightColor: 'rgba(155, 243, 22, 1)'
        }),
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