import 'umi/typings';

declare global {
namespace API {
    type EmptyData = Record<string, never>;

    type Response<T = EmptyData> = {
        code: number;
        message: string;
        data: T;
        request_id: string;
    };

    type RequestParams<TBody = any, TData = EmptyData> = {
        body?: TBody;
        onSuccess?: (res: Response<TData>) => void;
        onFail?: (res: Response<TData>) => void;
        onFinally?: () => void;
    };

    type LoginDevice = 'web' | 'pc' | 'mobile' | 'android';

    type Ui = {
        layout: string;
        watermark: boolean;
        theme: string;
        compact: boolean;
        color: string;
        radius: number;
    };

    type WebInfo = {
        title: string;
        name: string;
        ui: Ui;
    };

    type UserInfo = {
        account: string;
        login_ip: string;
        login_location: string;
        login_time: string;
    };
}
}

export {};
