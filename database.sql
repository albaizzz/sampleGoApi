
create table pet
(
	id int(9) not null auto_increment primary key,
    name varchar(250),
    age int(2),
    photo varchar(250)
)
select id, name, age, photo from pet where id =2
select id, name, age, photo from pet 
where id =1

delete from pet
where id =10

insert into pet  values(null, 'Ikan', 12,'photo')
usp_commissionMaster_GetByPeriod


delimiter |
create procedure usp_pet_insert (pname varchar(250), page int(2))
begin
	select * from pet;
end
|