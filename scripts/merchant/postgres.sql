-- INIT
\benchmark once \name initialize
DROP SCHEMA IF EXISTS godbbench CASCADE;
CREATE SCHEMA godbbench;
CREATE TABLE godbbench.customer (customer_id INT PRIMARY KEY, name VARCHAR(10), address VARCHAR(50), birthday  DATE);
CREATE TABLE "godbbench"."order" (order_id INT PRIMARY KEY, customer_id INT NOT NULL, creation_date DATE, comment VARCHAR(50), FOREIGN KEY (customer_id) REFERENCES godbbench.customer (customer_id));
CREATE TABLE godbbench.category (category_id INT PRIMARY KEY, name VARCHAR(10));
CREATE TABLE godbbench.supplier (supplier_id  INT PRIMARY KEY, name VARCHAR(10), address VARCHAR(50));
CREATE TABLE godbbench.product (product_id INT PRIMARY KEY, supplier_id INT NOT NULL, category_id INT NOT NULL, code VARCHAR(6), description VARCHAR(100), unit_size INT, price_per_unit DECIMAL(10,2), FOREIGN KEY (supplier_id) REFERENCES godbbench.supplier (supplier_id), FOREIGN KEY (category_id)  REFERENCES godbbench.category (category_id));
CREATE TABLE godbbench.line_item (line_item_id INT PRIMARY KEY, order_id INT NOT NULL, product_id INT NOT NULL, quantity INT, delivery_date DATE, FOREIGN KEY (order_id) REFERENCES "godbbench"."order" (order_id), FOREIGN KEY (product_id) REFERENCES godbbench.product (product_id));

-- INSERTS
\benchmark loop 1.0 \name inserts
INSERT INTO godbbench.customer (customer_id, name, address, birthday) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}', '{{call .RandDate }}');
INSERT INTO godbbench.order (order_id, customer_id, creation_date, comment) VALUES( {{.Iter}}, (SELECT customer_id FROM godbbench.customer ORDER BY RANDOM() LIMIT 1), '{{call .RandDate }}', '{{call .RandString 0 50 }}');
INSERT INTO godbbench.supplier (supplier_id, name, address) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}', '{{call .RandString 10 50 }}');
INSERT INTO godbbench.category (category_id, name) VALUES( {{.Iter}}, '{{call .RandString 3 10 }}');
INSERT INTO godbbench.product (product_id, supplier_id, category_id, code, description, unit_size, price_per_unit) VALUES( {{.Iter}}, (SELECT supplier_id FROM godbbench.supplier ORDER BY RANDOM() LIMIT 1), (SELECT category_id FROM godbbench.category ORDER BY RANDOM() LIMIT 1), '{{call .RandString 5 6 }}', '{{call .RandString 0 100 }}', {{call .RandIntBetween 1 10 }}, {{call .RandFloatBetween 0.01 999999.99 }});
INSERT INTO godbbench.line_item (line_item_id, order_id, product_id, quantity, delivery_date) VALUES( {{.Iter}}, (SELECT order_id FROM godbbench.order ORDER BY RANDOM() LIMIT 1), (SELECT product_id FROM godbbench.product ORDER BY RANDOM() LIMIT 1), {{call .RandIntBetween 1 5000 }}, '{{call .RandDate }}');

-- SELECTS
\benchmark loop 1.0 \name select_simple
SELECT * FROM godbbench.customer WHERE customer_id = {{.Iter}} 

\benchmark loop 1.0 \name select_medium
SELECT * FROM godbbench.product p JOIN godbbench.supplier s on p.supplier_id = s.supplier_id WHERE s.supplier_id = {{.Iter}} ORDER BY p.price_per_unit DESC
  
\benchmark loop 1.0 \name select_complex
SELECT c.customer_id, c.name, SUM(li.quantity * p.unit_size * p.price_per_unit) as TotalorderValue  FROM godbbench.customer c  INNER JOIN godbbench.order o on o.customer_id = c.customer_id  INNER JOIN godbbench.line_item li on o.order_id = li.order_id  INNER JOIN godbbench.product p on p.product_id = li.product_id  WHERE (o.creation_date BETWEEN '{{call .RandDate }}' AND '9999-12-31')  GROUP by c.customer_id, c.name  ORDER by c.customer_id

-- CLEAN
\benchmark once \name clean
DROP SCHEMA IF EXISTS godbbench CASCADE;
