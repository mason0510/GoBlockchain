pragma solidity ^0.4.0;

contract SimpleStorage {
    //正常的存取
    string[] stringData;
    function SimpleStorage(){

    }
    //定义数组 类似于数据库 输入输出 需不需要付钱
    function addData(string hash)public{
    stringData.push(hash);
    }
    //操作数组存取 查询数据库
    function getData(uint n)public view returns(uint,string){
        return (stringData.length,stringData[i]);
    }
}
