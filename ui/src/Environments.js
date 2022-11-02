import DataTable from 'react-data-table-component';
import Services from './Services';
import Icon from "@mdi/react";
import {mdiCloseOutline} from "@mdi/js";
import * as identity from './api/identity';

const Environments = (props) => {
    const humanizeDuration = require("humanize-duration")
    const disableEnvironment = (envId) => {
        console.log(envId)
        if(window.confirm('really disable environment "' + envId +'"?')) {
            identity.disable({body: {identity: envId}}).then(resp => {
                console.log(resp);
            })
        }
    }

    const columns = [
        {
            name: 'Description',
            selector: row => row.environment.description,
            sortable: true,
        },
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
            name: 'Identity',
            selector: row => row.environment.zId,
            sortable: true,
        },
        {
            name: 'Actions',
            selector: row => <>
                <button data-value={row.environment.zId} onClick={e => disableEnvironment(row.environment.zId)} title={"Disable Environment '"+row.environment.zId+"'"}>
                    <Icon path={mdiCloseOutline} size={0.7}/>
                </button>
            </>
        },
        {
            name: 'Uptime',
            selector: row => humanizeDuration(Date.now() - row.environment.updatedAt),
            sortable: true,
        },
    ]

    const servicesComponent = ({ data }) => <Services envId={data.environment.zId} services={data.services} />
    const servicesExpanded = row => row.services != null && row.services.length > 0

    return (
        <div>
            <h2>Environments</h2>
            { props.overview && props.overview.length > 0 && (
                <div>
                    <DataTable
                        columns={columns}
                        data={props.overview}
                        defaultSortFieldId={1}
                        expandableRows
                        expandableRowsComponent={servicesComponent}
                        expandableRowExpanded={servicesExpanded}
                    />
                </div>
            )}
        </div>
    )
};

export default Environments;
