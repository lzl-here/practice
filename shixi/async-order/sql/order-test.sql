create table if not exists order_action
(
    id          int auto_increment
        primary key,
    order_id    varchar(255) null,
    action_type varchar(255) null,
    app_id      varchar(255) null,
    constraint order_action_order_id_app_id_action_type_uindex
        unique (order_id, app_id, action_type)
);

