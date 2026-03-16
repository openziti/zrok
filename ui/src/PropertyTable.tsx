import React from "react";
import {camelToWords, objectToRows, PropertyRow} from "./model/util.ts";
import {Paper, Table, TableBody, TableCell, TableRow} from "@mui/material";

type PropertyTableProps = {
    object: Record<string, unknown> | null | undefined;
    custom?: Record<string, (row: PropertyRow) => React.ReactNode>;
    labels?: Record<string, string>;
}

const PropertyTable = ({ object, custom, labels }: PropertyTableProps) => {
    const data = objectToRows(object);

    const value = (row: PropertyRow) => {
        if(custom) {
            if(row.property in custom) {
                return custom[row.property](row);
            }
        }
        return row.value;
    }

    const renderLabel = (row: PropertyRow) => {
        if(labels) {
            if(row.property in labels) {
                return labels[row.property];
            }
        }
        return camelToWords(row.property);
    }

    return (
        <Paper>
            <Table>
                <TableBody>
                    {data.map((row) => (
                        <TableRow key={row.id}>
                            <TableCell sx={{ width: 100 }}><strong>{renderLabel(row)}</strong></TableCell>
                            <TableCell sx={{ width: 1000 }}>{value(row)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </Paper>
    );
}

export default PropertyTable;
