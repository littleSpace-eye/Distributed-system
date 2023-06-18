$(document).ready(function () {
    //查看数据库里所有数据
    $.ajax({
        url: "http://localhost:8080/products",
        type: "GET",
        success: function (response) {

            // // 使用for循环遍历数组中的每个对象
            for (var i = 0; i < response.length; i++) {
                var product = response[i];

                // 提取每个对象的属性值
                var id = product.ID;
                var name = product.Name;
                var num = product.Num;
                var style = product.Style;
                var provider = product.Provider;

                var row = '<tr>' +
                    '<th scope="row">' + id + '</th>' +
                    '<td>' + name + '</td>' +
                    '<td>' + num + '</td>' +
                    '<td>' + style + '</td>' +
                    '<td>' + provider + '</td>' +
                    '</tr>';

                //     // 将生成的行添加到表格中
                $('#product_list').append(row);
            }
        },
        error: function () {
            alert("请求失败！");
        }
    });
});


//添加库存
$("#add_product_btn").click(function () {
    let id = parseInt($("#add_product_id").val().trim());
    let quantity = parseInt($("#add_product_input").val().trim());

    console.log("id:", id, typeof (id), "quantity:", quantity, typeof (quantity))
    var data = {
        "id": id,
        "quantity": quantity

    };
    //执行AJAX Request POST请求
    $.ajax({
        url: "http://localhost:8080/product/add", // 请求的URL
        type: "POST", // 请求方法类型
        data: JSON.stringify(data), // 将JSON对象转换为JSON字符串
        contentType: "application/json",
        success: function (response) {
            alert("添加成功")
            // 执行页面刷新
            location.reload();
        },
        error: function () {
            alert("请求失败！"); // 在错误回调中显示警告框，包含输入框内容
            console.log(response)
        }

    });
});

//减少库存
$("#sub_product_btn").click(function () {
    let id = parseInt($("#sub_product_id").val().trim());
    let quantity = parseInt($("#sub_product_input").val().trim());

    console.log("id:", id, typeof (id), "quantity:", quantity, typeof (quantity))
    var data = {
        "id": id,
        "quantity": quantity

    };
    //执行AJAX Request POST请求
    $.ajax({
        url: "http://localhost:8080/product/sub", // 请求的URL
        type: "POST", // 请求方法类型
        data: JSON.stringify(data), // 将JSON对象转换为JSON字符串
        contentType: "application/json",
        success: function (response) {
            alert("减少成功")
            // 执行页面刷新
            location.reload();
        },
        error: function () {
            alert("请求失败！"); // 在错误回调中显示警告框，包含输入框内容
            console.log(response)
        }

    });
});


//通过id查询库存
$("#check_product_by_id_btn").click(function(){
    let id = parseInt($("#check_product_by_id").val().trim());
    var data ={
        "id":id
    }
    $.ajax({
        url:"http://localhost:8080/product/num",
        type:"POST",
        data:JSON.stringify(data),
        contentType: "application/json",
        success:function(response){
            var quantity = response.quantity;
            var quantity_displace = id+"号商品有"+quantity+"个库存";
            var quantity_container = document.getElementById("inventory_by_id");
            quantity_container.textContent = quantity_displace;
        },
        error:function(){
            console.log("查询失败",error)
        }

    })
})

//通过name查询库存
$("#check_product_by_name_btn").click(function(){
    let name = $("#check_product_by_name").val()
    var data ={
        "name":name
    }
    console.log(typeof(name))
     $.ajax({
        url:"http://localhost:8080/products/quantity/name",
        type:"POST",
        data:JSON.stringify(data),
        contentType: "application/json",
        success:function(response){
            console.log(response)
            var quantity = response.quantity;
            var quantity_displace = name+"商品有"+quantity+"个库存";
            var quantity_container = document.getElementById("inventory_by_name");
            quantity_container.textContent = quantity_displace;
        },
        error:function(){
            console.log("查询失败",error)
        }

    })
})