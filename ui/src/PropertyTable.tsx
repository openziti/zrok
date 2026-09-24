import {ReactNode, useEffect, useState} from "react";
import {camelToWords, objectToRows} from "./model/util.ts";
import {Paper, Table, TableBody, TableCell, TableRow} from "@mui/material";

type PropertyRow = {
    id: number;
    property: string;
    value: unknown;
}

type PropertyTableProps = {
    object: object;
    custom?: Record<string, (row: PropertyRow) => ReactNode>;
    labels?: Record<string, string>;
}

const PropertyTable = ({ object, custom, labels }: PropertyTableProps) => {
    const [data, setData] = useState([]);

    useEffect(() => {
        setData(objectToRows(object));
    }, [object]);

    const value = (row) => {
        if(custom) {
            if(row.property in custom) {
                return custom[row.property](row);
            }
        }
        return row.value;
    }

    const renderLabel = (row) => {
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