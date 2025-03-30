box.cfg({listen="0.0.0.0:3301"})

box.once("init", function()
    box.schema.user.create('storage', {password='pass', if_not_exists=true})
    box.schema.user.grant('storage', 'super', nil, nil, {if_not_exists=true})
    box.schema.sequence.create('voting_id_seq')
    s = box.schema.space.create('votings')
    s:format({
    {name = 'voting_id', type = 'unsigned'},
    {name = 'channel_id', type = 'string'},
    {name = 'voting_record', type = 'varbinary'}
    })

    s:create_index('idx_vt_id', {
        type = 'hash',
        parts = {'voting_id'},
        sequence = 'voting_id_seq'
        })

    s:create_index('idx_chan_id', {
        type = 'tree',
        parts = {'channel_id'},
        unique = false
        })
end)