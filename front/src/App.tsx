import {Button, Input, Tabs, Toast} from '@arco-design/mobile-react';
import '@arco-design/mobile-react/esm/style';
import React, {useState} from "react";
import backend from "./backend.tsx";

const tabData = [
    {title: 'token设置'},
    {title: '批量重命名'},
];

function App() {
    const [token, setToken] = React.useState('');
    const [url, setUrl] = React.useState('');
    const [name, setName] = React.useState('');
    const [prefix, setPrefix] = React.useState('');
    const [tokenLoading, setTokenLoading] = useState(false);
    const [nameLoading, setNameLoading] = useState(false);
    const [submitLoading, setSubmitLoading] = useState(false);
    const handleOpenToken = () => {
        window.open("https://alist.nn.ci/tool/aliyundrive/request.html", "_blank")
    }
    const handleGetName = async () => {
        setNameLoading(true);
        try {
            const response = await fetch(backend + `/v1/name`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    file_path: url,
                    name: name,
                    prefix: prefix,
                }),
            });
            if (response.ok) {
                const data = await response.json();
                setName(data.Name)
                Toast.toast('请求成功');
                // 处理你的数据
            } else {
                throw new Error('Something went wrong on api server!');
            }
        } catch (error) {
            Toast.toast('请求失败,请重试');
            console.error('Error occurred:', error);
        } finally {
            setNameLoading(false); // 关闭加载状态
        }
    }
    const handleSaveToken = async () => {
        setTokenLoading(true); // 开启加载状态
        try {
            const response = await fetch(backend + `/v1/token`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    token: token
                }),
            });
            if (response.ok) {
                Toast.toast('请求成功');
            } else {
                throw new Error('Something went wrong on api server!');
            }
        } catch (error) {
            Toast.toast('请求失败,请重试');
            console.error('Error occurred:', error);
        } finally {
            setTokenLoading(false); // 关闭加载状态
        }
    };
    const handleSubmit = async () => {
        setSubmitLoading(true);
        try {
            const response = await fetch(backend + `/v1/batch`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    file_path: url,
                    name: name,
                    prefix: prefix,
                }),
            });
            if (response.ok) {
                Toast.toast('请求成功');
                // 处理你的数据
            } else {
                throw new Error('Something went wrong on api server!');
            }
        } catch (error) {
            Toast.toast('请求失败,请重试');
            console.error('Error occurred:', error);
        } finally {
            setSubmitLoading(false); // 关闭加载状态
        }
    }
    return (
        <Tabs
            tabs={tabData}
            type="line-divide"
            defaultActiveTab={0}
            tabBarHasDivider={false}
            onAfterChange={(tab, index) => {
                console.log('[tabs]', tab, index);
            }}
            translateZ={false}
        >
            <div className="demo-tab-content">
                <Input
                    value={token}
                    onInput={(_, value) => setToken(value)}
                    label="Token"
                    placeholder="请输入新获取的阿里云盘Token"
                    clearable
                    onClear={() => {
                        setToken('');
                    }}
                    border="none"
                    clearShowType='always'
                />
                <div style={{display: 'flex', justifyContent: 'space-around'}}>
                    <Button type="default" needActive onClick={handleOpenToken}>获取Token</Button>
                    <Button needActive loading={tokenLoading} onClick={handleSaveToken}>保存Token</Button>
                </div>
            </div>
            <div className="demo-tab-content">
                <Input
                    value={url}
                    onInput={(_, value) => setUrl(value)}
                    label="目录地址"
                    placeholder="请输入需要批量修改的目录地址"
                    clearable
                    onClear={() => {
                        setUrl('');
                    }}
                    border="none"
                    clearShowType='always'
                />
                <div style={{display: 'flex', justifyContent: 'space-around', alignItems: 'center'}}>
                    <Input
                        style={{flex: '5'}}  // 调整为更大的比例
                        value={name}
                        onInput={(_, value) => setName(value)}
                        label="匹配格式"
                        placeholder="请输入如何匹配集数,如庆余年$01$.mp4"
                        clearable
                        onClear={() => {
                            setName('');
                        }}
                        border="none"
                        clearShowType='always'
                    />
                    <Button
                        style={{flex: '1'}}  // 调整为更小的比例
                        needActive
                        loading={nameLoading}
                        onClick={handleGetName}
                    >
                        填充
                    </Button>
                </div>
                <Input
                    value={prefix}
                    onInput={(_, value) => setPrefix(value)}
                    label="文件前缀"
                    placeholder="请输入新的文件前缀,如:庆余年S01"
                    clearable
                    onClear={() => {
                        setPrefix('');
                    }}
                    border="none"
                    clearShowType='always'
                />
                <Button
                    needActive
                    loading={submitLoading}
                    onClick={handleSubmit}
                >
                    批量修改
                </Button>
            </div>
        </Tabs>
    )
}

export default App
