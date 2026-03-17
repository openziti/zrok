import {Box} from "@mui/material";
import useApiConsoleStore from "./model/store.ts";
import {
    MaterialReactTable,
    type MRT_ColumnDef,
    MRT_PaginationState,
    MRT_RowSelectionState,
    MRT_SortingState,
    useMaterialReactTable
} from "material-react-table";
import {useEffect, useMemo, useState} from "react";
import {bytesToSize} from "./model/util.ts";
import {COLORS} from "./styling/theme.ts";

interface TableRow {
    id: string;
    label: string;
    type: string;
    activity?: number[];
}

const TabularView = () => {
    const nodes = useApiConsoleStore((state) => state.nodes);
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

    const rows = useMemo<TableRow[]>(() => {
        return nodes.map(node => ({
            id: node.id,
            label: String(node.data.label ?? node.id),
            type: node.type ?? "",
            activity: sparkdata.get(node.id),
        }));
    }, [nodes, sparkdata]);

    useEffect(() => {
        if(selectedNode) {
            const selection = {};
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
        const sn = nodes.find(node => Object.keys(rowSelection).includes(node.id));
        updateSelectedNode(sn ?? null);
        updateNodes(nodes.map(node => (sn && node.id === sn.id) ? { ...node, selected: true } : { ...node, selected: false }));
    }, [rowSelection]);

    const sparkdataTip = (row: TableRow): number => {
        if(row.activity) {
            // - 2; - 1 is sometimes undefined?
            return row.activity[row.activity.length - 2];
        }
        return 0;
    };

    const sparkdataTipFmt = (row: TableRow): string => {
        const tip = sparkdataTip(row);
        if(tip > 0) {
            return bytesToSize(tip);
        }
        return "";
    };

    const sparkdataAverage = (row: TableRow): number => {
        if(row.activity) {
            let average = row.activity.reduce((acc, curr) => { return acc + curr }, 0);
            average /= row.activity.length;
            return average;
        }
        return 0;
    };

    const sparkdataAverageFmt = (row: TableRow): string => {
        const average = sparkdataAverage(row);
        if(average > 0) {
            return bytesToSize(average);
        }
        return "";
    }

    const columns = useMemo<MRT_ColumnDef<TableRow>[]>(
        () => [
            {
                id: "data.label",
                accessorFn: row => row.label,
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
                    const tipA = sparkdataTip(rowA.original);
                    const tipB = sparkdataTip(rowB.original);
                    return tipA > tipB ? 1 : tipA < tipB ? -1 : 0;
                },
                sortDescFirst: true
            },
            {
                accessorFn: sparkdataAverageFmt,
                header: 'Activity 5m',
                sortingFn: (rowA, rowB) => {
                    const avgA = sparkdataAverage(rowA.original);
                    const avgB = sparkdataAverage(rowB.original);
                    return avgA > avgB ? 1 : avgA < avgB ? -1 : 0;
                },
                sortDescFirst: true
            }
        ],
        [],
    );

    const table = useMaterialReactTable({
        columns: columns,
        data: rows,
        enableStickyHeader: true,
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
                color: COLORS.primary,
                backgroundColor: COLORS.alertBannerBg,
            }
        },
        muiTableContainerProps: {
            sx: {
                flex: 1,
                minHeight: 0,
            }
        },
        muiTablePaperProps: {
            sx: {
                height: "100%",
                display: "flex",
                flexDirection: "column",
            }
        },
        positionToolbarAlertBanner: "bottom",
        mrtTheme: () => ({
            matchHighlightColor: COLORS.secondary
        }),
    });

    return (
        <Box sx={{ width: "100%", height: "100%", minHeight: 0 }}>
            <MaterialReactTable table={table} />
        </Box>
    );
};

export default TabularView;
