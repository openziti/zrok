import DataTable from 'react-data-table-component';

const Environments = (props) => {
    const columns = [
        {
            name: 'Host',
            selector: row => row.environment.host,
            sortable: true,
        },
        {
            name: 'Address',
            selector: row => row.environment.address,
            sortable: true,
        },
        {
            name: 'Description',
            selector: row => row.environment.description,
            sortable: true,
        },
        {
            name: 'Ziti Identity',
            selector: row => row.environment.zitiIdentityId,
            sortable: true,
        },
        {
            name: 'Active',
            selector: row => row.environment.active ? 'Active' : 'Inactive',
            sortable: true,
        },
    ]

    return (
        <div>
            <h2>Environments</h2>
            { props.overview && props.overview.length > 0 && (
                <div>
                    <DataTable
                        columns={columns}
                        data={props.overview}
                        defaultSortFieldId={1}
                    />
                </div>
            )}
        </div>
    )
};

export default Environments;
