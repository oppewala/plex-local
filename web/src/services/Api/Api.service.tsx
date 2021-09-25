import { Config } from "@utils/Config/config";
import {SearchResponse, DownloadResponse } from "@services/Api/types";

const Search = async (query: string, options?: RequestInit): Promise<Array<SearchResponse>> => {
    const url = new URL('/api/search', Config.ApiRoot)
    const opt: RequestInit = {
        ...options,
        method: 'GET',
    };
    url.searchParams.append("q", query);

    const res = await fetch(url.toString(), opt);
    const j: Array<SearchResponse> = await res.json();

    return j
}

const Download = async (key: string, options?: RequestInit): Promise<DownloadResponse> => {
    const path = '/api/media/' + key + '/download';
    const url = new URL(path, Config.ApiRoot)
    const opt: RequestInit = {
        ...options,
        method: 'POST',
    };

    const res = await fetch(url.toString(), opt);
    const j: DownloadResponse = await res.json();

    return j
}

export { Search, Download };