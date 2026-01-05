//initialize socket and message storage array
let socket = null;
const messages = [];
let ping_interval = 45000; //ping interval 45 seconds

//connection elements
const userId_input = document.getElementById("userId_input"); //id input
const auto_reconnect_checkbox = document.getElementById("auto_reconnect_checkbox"); //auto reconnect checkbox
const connection_status_label = document.getElementById("connection_status_label"); //connection status label
const connectButton = document.getElementById("connectButton"); //connection button

//send message elements
const message_type_input = document.getElementById("message_type_input");
const target_Id_input = document.getElementById("target_Id_input");
const message_content_input = document.getElementById("message_content_input");
const base64_encoding_checkBox = document.getElementById("base64_encoding_checkBox");
const send_button = document.getElementById("send_button");

//receive message elements
const message_list = document.getElementById("message_list");

//detailed view modal elements
const modal_msg_type = document.getElementById("modal_msg_type");
const modal_msg_from = document.getElementById("modal_msg_from");
const modal_msg_content = document.getElementById("modal_msg_content");
const modal_msg_time = document.getElementById("modal_msg_time");
const modal_msg_raw = document.getElementById("modal_msg_raw");
const modal_msg_to = document.getElementById("modal_msg_to");

//decode base64
const base_64_decoder_button = document.getElementById("base_64_decoder_button");


//sed ping message to keep client alive 
function ping_socket(){
    if(socket!=null){
        socket.ping();
    }
}

//detailed view message modal
function openMore(number){
    const selected_message = messages[number]; //get selected message
    //validate message
    if(selected_message != undefined){
        modal_msg_type.value = selected_message.msg_type; //message type
        modal_msg_from.value = selected_message.from; //message from
        modal_msg_content.value = selected_message.message; //message content
        modal_msg_time.value = selected_message.time; //message time
        modal_msg_to.value = selected_message.to; //message to
        modal_msg_raw.value = JSON.stringify(selected_message); //message raw
    }
}

//refresh the messages in message list
function refresh_messages(){
    let all_messages = ``;
    messages.forEach((element,index)=>{
        max_length=5;
        const message_type = element.msg_type.length<max_length?element.msg_type:element.msg_type.slice(0,max_length)+"..."; //replace excess chars with ...
        const message_content = element.message.length<max_length?element.message:element.message.slice(0,max_length)+"...";//replace excess chars with ...
        const from  = element.from.length<max_length?element.from:element.from.slice(0,max_length)+"..."; //replace excess chars with ...
        all_messages+=`<li data-bs-toggle="modal" data-bs-target="#detailModal" onclick="openMore(${index})" class="list-group-item message_item"><div class="time_element">${message_type}</div><div class="message_element">${message_content}</div><div class="client_element">${from}</div></li>`
    });

    message_list.innerHTML = ``;
    message_list.innerHTML = all_messages; //innerhtml not safe BTW :(    
}

//manage socket events
function listen_sockets(){
    //on socket connected
    socket.socket.addEventListener('open',()=>{
        connection_status_label.textContent = `Connected`;
        connection_status_label.style.color = "#5b5383";

        //start ping interval
        if(auto_reconnect_checkbox.checked){
            console.log("Alive Connection : Enabled")
            setInterval(ping_socket, ping_interval);
        }
        
    });
    
    //on socket message
    socket.socket.addEventListener('message',(event)=>{
        const json_data = JSON.parse(`${event.data}`)
        json_data["time"]=new Date().toLocaleString();
        messages.push(json_data);
        refresh_messages();
    });

    //on socket disconnected
    socket.socket.addEventListener('close',()=>{
        connection_status_label.textContent = `Disconnected`;
        connection_status_label.style.color = "#acacae";
        if(auto_reconnect_checkbox.checked){
            socket.reconnect();
            listen_sockets();
        }
    });

}

//connect button event
connectButton.addEventListener('click',()=>{
    if(userId_input.value!=""){
        socket = new Fsocket(userId_input.value); //connect to socket
        listen_sockets(); //start listening to events
    }
});

send_button.addEventListener('click',()=>{
    //validate before sending
    if(socket!=null && message_type_input.value !="" && message_content_input.value !="" && target_Id_input.value !=""){
        const message_content = base64_encoding_checkBox.checked?btoa(message_content_input.value):message_content_input.value; //encode base64 if checked
        socket.sendMessage(message_type_input.value,message_content,target_Id_input.value); //send message
        
        //add SEND message to message array
        const json_data = {msg_type:message_type_input.value,from:"YOU",message:message_content,time:new Date().toLocaleString(),to:target_Id_input.value}
        messages.push(json_data);
        refresh_messages(); //refresh messages
    }
});

//decode base 64
base_64_decoder_button.addEventListener('click',()=>{
    modal_msg_content.value = atob(modal_msg_content.value);
});