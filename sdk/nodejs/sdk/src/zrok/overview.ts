import {Root} from "../environment/root"
import get, {AxiosHeaders} from "axios"

export async function Overview(root: Root): Promise<string> {
    if (!root.IsEnabled()){
        throw new Error("environment is not enabled; enable with 'zrok enable' first!")
    }
    let headers = new AxiosHeaders()
    headers.set("X-TOKEN", root.env.Token)

    let resp = await get(root.ApiEndpoint().endpoint + "/api/v1/overview", {headers: headers})
    let data = JSON.parse(resp.data)
    return data
}