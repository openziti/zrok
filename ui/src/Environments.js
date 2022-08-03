import {useEffect, useState} from "react";
import * as metadata from './api/metadata';
import DataTable from 'react-data-table-component';

const Environments = (props) => {
    const [environments, setEnvironments] = useState([])

    const columns = [
        {
            name: 'Ziti Identity',
            selector: row => row.zitiIdentityId,
            sortable: true,
        },
        {
            name: 'Active',
            selector: row => row.active ? 'Active' : 'Inactive',
            sortable: true,
        },
        {
            name: 'Created At',
            selector: row => row.createdAt,
            sortable: true,
        },
        {
            name: 'Updated At',
            selector: row => row.updatedAt,
            sortable: true,
        }
    ]

    useEffect(() => {
        let mounted = true;
        metadata.listEnvironments().then((resp) => {
            setEnvironments(resp.data)
            console.log(resp.data);
        })
        return () => { mounted = false; }
    }, [])

    return (
        <div>
            <h1>Environments</h1>
            { environments && environments.length > 0 && (
                <div>
                    <DataTable
                        columns={columns}
                        data={environments}
                        defaultSortFieldId={1}
                    />
                </div>
            )}
        </div>
    )
};

export default Environments;
