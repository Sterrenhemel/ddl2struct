CREATE TABLE ddl2struct
(
    person_id  bigint comment 'ID',
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