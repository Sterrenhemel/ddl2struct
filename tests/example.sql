CREATE TABLE ddl2struct
(
    person_id  bigint comment 'ID',
    it         integer comment 'id',
    tit        tinyint comment 'tinyint',
    last_name  varchar(255) comment 'last Name',
    first_name varchar(255) comment 'first Name',
    address    varchar(255) comment 'address',
    city       varchar(255) comment 'city'
) comment ' aaa.go';

CREATE TABLE ddl2struct2
(
    person_id  bigint,
    last_name  varchar(255),
    first_name varchar(255),
    address    varchar(255),
    city       varchar(255)
);

create table admin_role
(
    id          bigint       null comment '唯一id',
    role_key    varchar(255) null comment '角色key',
    description varchar(255) null comment '角色描述',
    status      boolean         null comment '角色状态',
    name        varchar(255) null comment '角色名称',
    primary key admin_role_pk(id),
    unique(role_key)
) comment '北极星权限角色表';

create table admin_role_permission_relation
(
    role_id       bigint null comment '角色id',
    permission_id bigint null comment '权限id',
    constraint role_permission_uk
        unique (role_id, permission_id)
) comment '北极星角色权限关联表';