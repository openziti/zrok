import {useEffect, useState} from "react";
import * as metadata from './api/metadata';
import DataTable from 'react-data-table-component';

const Identities = (props) => {
    const [identities, setIdentities] = useState([])

    const columns = [
        {
            name: 'Ziti ID',
            selector: row => row.zitiId,
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
        metadata.listIdentities().then((resp) => {
            setIdentities(resp.data)
            console.log(resp.data);
        })
        return () => { mounted = false; }
    }, [])

    return (
        <div>
            <h1>Identities</h1>
            { identities && identities.length > 0 && (
                <div>
                    <DataTable
                        columns={columns}
                        data={identities}
                        defaultSortFieldId={1}
                    />
                </div>
            )}
        </div>
    )
};

export default Identities;
