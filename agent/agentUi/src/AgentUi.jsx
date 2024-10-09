import {createBrowserRouter, RouterProvider} from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";
import {useEffect, useState} from "react";
import {AgentApi, ApiClient} from "./api/src/index.js";

const AgentUi = (props) => {
    const [version, setVersion] = useState("");

    let api = new AgentApi(new ApiClient(window.location.protocol+'//'+window.location.host));

    useEffect(() => {
        let mounted = true;
        api.agentVersion((err, data) => {
            if(mounted) {
                setVersion(data.v);
            }
        });
    }, []);

    const router = createBrowserRouter([
        {
            path: "/",
            element: <Overview version={version} />
        },
        {
            path: "/share/:token",
            element: <ShareDetail version={version} />
        }
    ]);

    return <RouterProvider router={router} />
}

export default AgentUi;