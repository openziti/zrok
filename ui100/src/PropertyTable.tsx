import React, {useEffect, useState} from "react";
import {camelToWords, objectToRows} from "./model/util.ts";
import {Paper, Table, TableBody, TableCell, TableRow} from "@mui/material";

type PropertyTableProps = {
    object: any;
    custom: any;
}

const PropertyTable = ({ object, custom }: PropertyTableProps) => {
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

    return (
        <Paper>
            <Table>
                <TableBody>
                    {data.map((row) => (
                        <TableRow key={row.id}>
                            <TableCell sx={{ width: 100 }}><strong>{camelToWords(row.property)}</strong></TableCell>
                            <TableCell sx={{ width: 1000 }}>{value(row)}</TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </Paper>
    );
}

export default PropertyTable;