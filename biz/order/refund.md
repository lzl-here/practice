┌──────────────┐   ┌─────────────┐   ┌───────────────┐
│ 退款申请       │──▶│ 规则引擎      │──▶│ 自动审核队列    │（优先级分桶）
└───────┬──────┘   └──────┬──────┘   └──────┬────────┘
        │                 │                 │
┌───────▼──────┐   ┌──────▼──────┐   ┌──────▼────────┐
│ 用户历史行为分析 │  │ 风控评分模型   │  │ 人工审核队列    │（高风险/复杂场景）
└──────────────┘   └─────────────┘   └───────────────┘

``` java
// Drools规则示例
rule "Auto-Approval-L1" 
    when
        $r : RefundRequest(amount < 500, 
                           user.creditLevel > 3,
                           createTime > 30天前)
    then
        $r.approve();
end
```

跨平台订单聚合调度
``` text
            ┌───────────────┐
            │ 订单特征提取    │
            │ - 商品类目     ｜
            │ - 配送地址     │
            │ - 时效要求     │
            └───────┬───────┘
                    │
            ┌───────▼───────┐   ┌─────────┐
            │ 成本最优决策树  │──▶│ 本地仓储 │
            │               │   └─────────┘
            │               ├──▶│ 区域中心仓 │
            │               │   └──────────┘
            └───────────────┘         ▲
                                 ┌────┴────┐
                                 │ 第三方物流 │
                                 └─────────┘
```

