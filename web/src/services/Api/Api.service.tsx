import {Config} from "@utils/Config/config";
import {DownloadResponse, SearchResponse} from "@services/Api/types";

const Search = async (query: string, options?: RequestInit): Promise<Array<SearchResponse>> => {
    const url = new URL('/api/search', Config.ApiRoot)
    const opt: RequestInit = {
        ...options,
        method: 'GET',
    };
    url.searchParams.append("q", query);

    const res = await fetch(url.toString(), opt);
    return await res.json()
}

const Download = async (key: string, downloadFuture: boolean, options?: RequestInit): Promise<DownloadResponse> => {
    const path = '/api/media/' + key + '/download';
    const url = new URL(path, Config.ApiRoot)
    const opt: RequestInit = {
        ...options,
        method: 'POST'
    };

    const qs = downloadFuture ? '?a=true' : '?a=false'

    const res = await fetch(`${url.toString()}${qs}`, opt);
    return await res.json()
}

export { Search, Download };