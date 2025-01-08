import {Box, Paper} from "@mui/material";
import useStore from "./model/store.ts";
import {MaterialReactTable, type MRT_ColumnDef, useMaterialReactTable} from "material-react-table";
import {useMemo} from "react";
import {Node} from "@xyflow/react";

const data: Node[] = [];

const TabularView = () => {
    const overview = useStore((state) => state.overview);

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
        data: overview.nodes,
    });

    console.log(overview.nodes);

    return (
        <Box sx={{ width: "100%", mt: 2 }} height={{ xs: 400, sm: 600, md: 800 }}>
            <Paper>
                <MaterialReactTable table={table} />
            </Paper>
        </Box>
    );
};

export default TabularView;